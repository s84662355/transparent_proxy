package tProxy

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"transparent/gvisor.dev/gvisor/pkg/tcpip"
	"transparent/gvisor.dev/gvisor/pkg/tcpip/header"
	"transparent/gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"transparent/gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"transparent/gvisor.dev/gvisor/pkg/tcpip/stack"
	"transparent/gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"transparent/log"

	"github.com/lysShub/divert-go"
	//"go.uber.org/zap"
)

// init 初始化函数，加载WinDivert驱动
func init() {
	defer log.Recover("divert MustLoad")
	divert.MustLoad(divert.DLL) // 确保WinDivert驱动加载
}

// initProxyServer 初始化代理服务器
func (m *manager) initProxyServer() error {
	// 1. 获取网络接口信息
	IfIdx, SubIfIdx, mtu, err := GetInterfaceIndex()
	if err != nil {
		// log.Panic("获取网络接口索引失败", zap.Error(err))
		return fmt.Errorf("获取网络接口索引失败 error:%w", err)
	}

	// 3. 设置WinDivert过滤器
	filter := fmt.Sprintf("ifIdx = %d and ip and tcp and outbound and not loopback  ", IfIdx)
	handle, err := divert.Open(filter, divert.Network, -1000, 0)
	if err != nil {
		return fmt.Errorf("打开WinDivert句柄失败 filter:%s error:%w", filter, err)
	}

	// 4. 保存网络配置
	m.ifIdx = IfIdx
	m.subIfIdx = SubIfIdx
	m.handle = handle
	m.mtu = mtu

	// 5. 配置默认地址参数
	addrrr := &divert.Address{}
	addrrr.Layer = divert.Network
	addrrr.Event = divert.NetworkPacket

	nw := addrrr.Network()
	nw.IfIdx = m.ifIdx
	nw.SubIfIdx = m.subIfIdx

	addrrr.SetIPv6(false)        // 仅IPv4
	addrrr.SetOutbound(false)    // 入站流量
	addrrr.SetLoopback(false)    // 非回环
	addrrr.SetIPChecksum(true)   // 校验IP校验和
	addrrr.SetTCPChecksum(false) // 不校验TCP校验和
	addrrr.SetImpostor(false)    // 非伪造包
	addrrr.SetUDPChecksum(false) // 不校验UDP校验和

	m.defaultAddrrr = addrrr

	return nil
}

// createStack 创建网络协议栈
func (m *manager) createStack() error {
	const NICID = tcpip.NICID(1)

	// 1. 创建新协议栈
	s := stack.New(stack.Options{
		NetworkProtocols:   []stack.NetworkProtocolFactory{ipv4.NewProtocol},
		TransportProtocols: []stack.TransportProtocolFactory{tcp.NewProtocol},
	})

	// 2. 创建通道端点
	channelEp := channel.New(512, m.mtu, "")
	channelEp.LinkEPCapabilities |= stack.CapabilityRXChecksumOffload // 禁用校验和检查
	ep := stack.LinkEndpoint(channelEp)

	// 3. 创建NIC
	if tcperr := s.CreateNIC(NICID, ep); tcperr != nil {
		channelEp.Close() // 关闭端点
		s.Destroy()
		// log.Panic("创建NIC失败", zap.Any("error", tcperr))
		return fmt.Errorf("创建NIC失败 error:%+v", tcperr)
	}

	// 4. 设置混杂模式
	if tcperr := s.SetPromiscuousMode(NICID, true); tcperr != nil {
		channelEp.Close() // 关闭端点
		s.Destroy()
		return fmt.Errorf("设置混杂模式失败 error:%+v", tcperr)
		// log.Panic("设置混杂模式失败", zap.Any("error", tcperr))
	}

	// 5. 设置欺骗模式
	if tcperr := s.SetSpoofing(NICID, true); tcperr != nil {
		channelEp.Close() // 关闭端点
		s.Destroy()
		return fmt.Errorf("设置欺骗模式失败 error:%+v", tcperr)

		// log.Panic("设置欺骗模式失败", zap.Any("error", tcperr))
	}

	// 6. 设置路由表
	s.SetRouteTable([]tcpip.Route{
		{
			Destination: header.IPv4EmptySubnet,
			NIC:         NICID,
		},
	})

	m.maxInFlight = 1 << 15

	m.tcpForwarder = tcp.NewForwarder(
		s,
		16<<10,                     // 16KB接收缓冲区
		m.maxInFlight,              // 32KB最大段大小
		m.transportProtocolHandler, // 自定义处理函数
	)

	// 7. 设置TCP协议处理器
	s.SetTransportProtocolHandler(
		tcp.ProtocolNumber,
		m.HandleTcpPacket,
	)

	m.channelEp = channelEp
	m.tcpipStack = s

	return nil
}

// closeDev 关闭网络设备
func (m *manager) closeDev() (error, error) {
	return m.handle.Shutdown(divert.Both), m.handle.Close()
}

// GetInterfaceIndex 获取网络接口索引信息
func GetInterfaceIndex() (uint32, uint32, uint32, error) {
	// 1. 定义DNS查询过滤规则
	const filter = "  not loopback and outbound and (ip.DstAddr = 8.8.8.8 or ipv6.DstAddr = 2001:4860:4860::8888) and tcp.DstPort = 53"

	// 2. 打开WinDivert句柄
	hd, err := divert.Open(filter, divert.Network, 0, divert.Sniff)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("打开WinDivert失败: %w", err)
	}
	defer hd.Close()
	defer hd.Shutdown(divert.Both)

	// 3. 启动并发DNS查询
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	startConnection(wg, "tcp4", "8.8.8.8:53")                // IPv4查询
	startConnection(wg, "tcp6", "[2001:4860:4860::8888]:53") // IPv6查询

	// 4. 接收数据包
	addr := divert.Address{}
	buff := make([]byte, 1500)

	errChan := make(chan error, 1)
	defer func() {
		for range errChan {
		}
	}()
	go func() {
		defer close(errChan)
		_, err = hd.Recv(buff, &addr)
		errChan <- err
	}()

	// 5. 设置超时
	ctxTimeOut, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	select {
	case <-ctxTimeOut.Done():
		return 0, 0, 0, fmt.Errorf("操作超时")
	case err := <-errChan:
		if err != nil {
			return 0, 0, 0, err
		}
	}

	// 6. 获取MTU信息
	nw := addr.Network()
	interfaces, err := net.Interfaces()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("获取网络接口失败: %w", err)
	}

	var mtu uint32
	for _, iface := range interfaces {
		if iface.Index == int(nw.IfIdx) {
			mtu = uint32(iface.MTU)
			break
		}
	}

	return nw.IfIdx, nw.SubIfIdx, mtu, nil
}

// startConnection 启动网络连接
func startConnection(wg *sync.WaitGroup, network, address string) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := net.DialTimeout(network, address, time.Second)
		if err != nil {
			return
		}
		conn.Close()
	}()
}
