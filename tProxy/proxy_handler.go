package tProxy

import (
	"fmt"
	"io"
	"net"
	"sync"

	"go.uber.org/zap"
	"transparent/gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"transparent/gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"transparent/gvisor.dev/gvisor/pkg/waiter"
	//"time"
	"transparent/log"
	"transparent/proto/http"
	"transparent/proto/oks"
	"transparent/proto/socks"
	"transparent/proto/trojan"
)

// transportProtocolHandler 处理TCP转发请求
// 参数r是TCP转发请求对象
func (m *manager) transportProtocolHandler(r *tcp.ForwarderRequest) {
	// 创建等待队列，用于TCP端点的异步操作
	wq := waiter.Queue{}

	// 创建TCP端点(Endpoint)
	ep, tcperr := r.CreateEndpoint(&wq)
	if tcperr != nil {
		// log.Error("创建TCP端点失败", zap.Any("error", tcperr))
		r.Complete(true) // 标记请求完成(带错误)
		return
	}
	defer ep.Close()        // 确保函数退出时关闭端点
	defer r.Complete(false) // 标记请求成功完成

	id := r.ID()

	addr := fmt.Sprintf("%s:%d", id.LocalAddress.String(), id.LocalPort)

	// 获取到目标地址的连接
	target, err := m.getConn(addr)
	if err != nil {
		return
	}
	defer target.Close() // 确保函数退出时关闭目标连接

	// 将gVisor的TCP端点包装为Go标准的net.Conn接口
	cep := gonet.NewTCPConn(&wq, ep)
	defer cep.Close() // 确保函数退出时关闭连接

	// 创建错误通道，用于协程间通信
	errChan := make(chan error, 2)
	defer close(errChan) // 确保函数退出时关闭通道

	// 使用WaitGroup等待两个协程完成
	wg := &sync.WaitGroup{}
	// 等待所有协程完成
	defer wg.Wait()

	wg.Add(2) // 需要等待2个协程

	// 启动协程1：从目标连接读取数据并写入客户端端点
	go func() {
		defer wg.Done()
		_, err := io.Copy(cep, target)
		errChan <- err // 发送可能发生的错误
	}()

	// 启动协程2：从客户端端点读取数据并写入目标连接
	go func() {
		defer wg.Done()
		_, err = io.Copy(target, cep)
		errChan <- err // 发送可能发生的错误
	}()

	// 等待以下两种情况之一发生：
	select {
	case err := <-errChan: // 1. 任一协程发生错误
		if err != nil {
			log.Error("数据传输错误", zap.Any("error", err))
		}
	case <-m.tcm.Context().Done(): // 2. 上下文被取消

	}

	// 关闭连接
	cep.Close()
	target.Close()

	fmt.Println("连接代理结束")
}

// getConn 根据配置获取到目标地址的连接
// 返回net.Conn连接对象和可能的错误
func (m *manager) getConn(addr string) (net.Conn, error) {
	switch m.proxyJson.ProxyType {
	case "socks": // SOCKS代理
		return socks.GetConn(m.tcm.Context(), m.proxyJson.ProxyUrl, addr)
	case "http": // HTTP代理
		return http.GetConn(m.tcm.Context(), m.proxyJson.ProxyUrl, addr)
	// Trojan代理支持，
	case "trojan":
		return trojan.GetConn(
			m.tcm.Context(),
			m.proxyJson.TrojanProxy.Server,
			m.proxyJson.TrojanProxy.Password,
			addr,
			m.proxyJson.TrojanProxy.InsecureSkipVerify,
		)
	case "oks":
		return oks.GetConn(m.tcm.Context(), m.proxyJson.ProxyUrl, addr)
	default: // 不支持的代理类型

	}

	// 创建 Dialer 实例
	dialer := &net.Dialer{}

	// 使用上下文进行连接
	return dialer.DialContext(m.tcm.Context(), "tcp", addr)
}
