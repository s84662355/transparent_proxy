package tProxy

import (
	"context"
	"fmt"
	"os"

	"github.com/lysShub/divert-go"
	"github.com/pkg/errors"
	net2 "github.com/shirou/gopsutil/net"
	"golang.org/x/sys/windows"

	"transparent/gvisor.dev/gvisor/pkg/buffer"
	"transparent/gvisor.dev/gvisor/pkg/tcpip/header"
	"transparent/gvisor.dev/gvisor/pkg/tcpip/stack"
)

// runReadDivert 从Windows Divert驱动读取并处理网络数据包
// ctx: 上下文对象，用于控制协程生命周期和取消信号
func (m *manager) runReadDivert(ctx context.Context) {
	var addr divert.Address        // 存储数据包的目标地址信息
	tempBuf := make([]byte, m.mtu) // 创建临时缓冲区，大小为MTU

Loop: // 主循环标签
	for {
		// 从Divert驱动读取数据包
		n, err := m.handle.Recv(tempBuf, &addr)
		if err != nil {
			if errors.Is(err, windows.ERROR_INSUFFICIENT_BUFFER) {
				// 缓冲区不足，跳过当前数据包继续循环
				goto Loop
			} else {
				// 其他错误直接返回
				return
			}
		} else if n == 0 {
			// 读取到空数据包，继续循环
			goto Loop
		}

		// 创建数据包副本以供goroutine安全使用
		packetCopy := make([]byte, n)
		copy(packetCopy, tempBuf[:n])
		addrCopy := addr // 复制地址结构体

		// 将数据包处理任务提交到任务管理器异步执行
		m.handlePacket(ctx, packetCopy, &addrCopy)

	}
}

// handlePacket 处理单个网络数据包
// packet: 数据包字节切片
// addr: 数据包的目标地址信息
func (m *manager) handlePacket(ctx context.Context, packet []byte, addr *divert.Address) {
	// 1. 基本长度检查
	if len(packet) < header.IPv4MinimumSize {
		return
	}

	// 2. IPv4头部验证
	ipv4 := header.IPv4(packet)
	if !ipv4.IsValid(len(packet)) {
		// 无效IP包，直接原样转发
		m.handle.Send(packet, addr)
		return
	}

	// 5. TCP头部验证
	tcpHdr := header.TCP(ipv4.Payload())

	if tcpHdr.Flags().Contains(header.TCPFlagSyn) && !tcpHdr.Flags().Contains(header.TCPFlagAck) {
		if conns, err := net2.ConnectionsWithContext(ctx, "tcp4"); err == nil && len(conns) > 0 {
			// 提取连接四元组信息
			srcAddr := ipv4.SourceAddress().String()
			srcPort := tcpHdr.SourcePort()
			dstAddr := ipv4.DestinationAddress().String()
			dstPort := tcpHdr.DestinationPort()
			// 存储源地址

			connKey := fmt.Sprintf("%s:%d:%s:%d", srcAddr, srcPort, dstAddr, dstPort)
			ppid := int32(os.Getpid())

			for _, conn := range conns {
				if connKey == fmt.Sprintf("%s:%d:%s:%d",
					conn.Laddr.IP, conn.Laddr.Port, conn.Raddr.IP, conn.Raddr.Port) {
					if ppid == conn.Pid {
						m.handle.Send(packet, addr)
						return
					}
				}
			}

			m.tTLMap.Set(connKey)
			m.handleProxyConnection(packet)
		}

		return
	}

	// 提取连接四元组信息
	srcAddr := ipv4.SourceAddress().String()
	srcPort := tcpHdr.SourcePort()
	dstAddr := ipv4.DestinationAddress().String()
	dstPort := tcpHdr.DestinationPort()
	// 存储源地址
	connKey := fmt.Sprintf("%s:%d:%s:%d", srcAddr, srcPort, dstAddr, dstPort)
	if m.tTLMap.Get(connKey) {
		m.handleProxyConnection(packet)
		// m.handle.Send(packet, addr)
		return
	}
	m.handle.Send(packet, addr)
}

// handleProxyConnection 处理代理连接的数据包
// packet: 已验证的TCP数据包
func (m *manager) handleProxyConnection(packet []byte) {
	// 将原始数据包转换为gVisor的PacketBuffer格式
	pktBuffer := stack.NewPacketBuffer(stack.PacketBufferOptions{
		Payload: buffer.MakeWithData(packet),
	})

	// 将数据包注入到gVisor网络栈中
	m.channelEp.InjectInbound(header.IPv4ProtocolNumber, pktBuffer)

	// 减少引用计数（gVisor内存管理机制）
	pktBuffer.DecRef()
}
