package tProxy

import (
	"context"

	//"go.uber.org/zap"
	"transparent/gvisor.dev/gvisor/pkg/tcpip/header"
	"transparent/gvisor.dev/gvisor/pkg/tcpip/stack"
	//"transparent/log"
)

// runReadStack 持续从网络栈读取数据包并进行处理
// ctx: 上下文对象，用于控制协程生命周期和取消信号
func (m *manager) runReadStack(ctx context.Context) {
	// 无限循环读取数据包，直到上下文取消或读取失败
	for {
		// 带上下文的读取操作，允许被取消
		pkt := m.channelEp.ReadContext(ctx)
		// 也可以使用不带上下文的读取: pkt := m.channelEp.Read()

		// 检查数据包有效性
		if pkt != nil && pkt.Size() > 0 {
			// 将数据包处理任务提交到任务管理器异步执行
			m.processPkt(pkt)
		}

		// 如果读取到nil，表示连接已关闭或出错，退出循环
		if pkt == nil {
			return
		}
	}
}

// processPkt 处理单个网络数据包
// pkt: 待处理的数据包指针
func (m *manager) processPkt(pkt *stack.PacketBuffer) {
	// 只处理TCP协议的数据包
	if pkt.TransportProtocolNumber == header.TCPProtocolNumber {
		// 创建缓冲区，大小为数据包总大小
		buf := make([]byte, pkt.Size(), pkt.Size())

		// 将数据包各部分拷贝到缓冲区:
		// 1. 拷贝网络层头部(如IP头)
		copy(buf[:], pkt.NetworkHeader().Slice())
		// 2. 拷贝传输层头部(如TCP头)
		copy(buf[len(pkt.NetworkHeader().Slice()):], pkt.TransportHeader().Slice())
		// 3. 拷贝实际负载数据
		copy(buf[len(pkt.NetworkHeader().Slice())+len(pkt.TransportHeader().Slice()):],
			pkt.Data().AsRange().ToSlice())

		// 通过handle发送处理后的数据
		m.handle.Send(buf, m.defaultAddrrr)
		// _, err := m.handle.Send(buf, m.defaultAddrrr)
		// if err != nil {
		// 	log.Error("发送处理后的数据包失败",
		// 		zap.Any("error", err),
		// 		zap.String("function", "processPkt handle send"))
		// }
	}
}
