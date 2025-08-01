package tProxy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lysShub/divert-go" // Windows 网络数据包捕获库

	"transparent/gvisor.dev/gvisor/pkg/tcpip/stack"
	"transparent/gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	//	"go.uber.org/zap"                          // 高性能日志库
	"transparent/gvisor.dev/gvisor/pkg/tcpip/link/channel" // gVisor 的网络栈实现
	//"transparent/log"

	"transparent/utils/taskConsumerManager"
)

type Manager interface {
	Start() (<-chan error, error)
	Stop()
}

// manager 结构体管理整个代理服务的核心组件
type manager struct {
	tcm               *taskConsumerManager.Manager // 任务调度管理器
	handle            *divert.Handle               // WinDivert 句柄，用于网络包捕获
	channelEp         *channel.Endpoint            // gVisor 网络栈的端点
	tcpipStack        *stack.Stack
	ifIdx             uint32 // 网络接口索引
	subIfIdx          uint32 // 子接口索引
	mtu               uint32 // 最大传输单元
	defaultAddrrr     *divert.Address
	exitChan          chan error
	exitChanCloseFunc func()
	tcpForwarder      *tcp.Forwarder
	mu                sync.Mutex
	inFlight          map[stack.TransportEndpointID]struct{}
	maxInFlight       int
	thwg              sync.WaitGroup
	channelEpClose    func()
	proxyJson         *ProxyJson
	tTLMap            *TTLMap
	start             func() (<-chan error, error)
	stop              sync.Once
}

// 使用 sync.OnceValue 确保 manager 只被初始化一次（线程安全）
func NewManager(proxyJson *ProxyJson) *manager {
	m := &manager{
		tcm:       taskConsumerManager.New(), // 任务消费者管理器
		exitChan:  make(chan error, 1),
		proxyJson: proxyJson,
	}
	m.exitChanCloseFunc = sync.OnceFunc(func() {
		close(m.exitChan)
	})

	m.start = sync.OnceValues[<-chan error, error](func() (<-chan error, error) {
		select {
		case <-m.tcm.Context().Done():
			return nil, fmt.Errorf("已关闭")
		default:
		}

		// 初始化代理服务器
		if err := m.initProxyServer(); err != nil {
			return nil, err
		}

		// 创建网络协议栈
		if err := m.createStack(); err != nil {
			m.closeDev() // 关闭设备
			return nil, err
		}

		m.channelEpClose = sync.OnceFunc(func() {
			if m.channelEp != nil {
				m.channelEp.Close()
			}
		})

		m.tTLMap = NewTTLMap(30 * time.Second)
		m.tcm.AddTask(1, func(ctx context.Context) {
			m.tTLMap.Start()
			defer m.tTLMap.Stop()
			<-ctx.Done()
		})

		// 添加三个并行运行的守护任务：
		m.tcm.AddTask(1, func(ctx context.Context) {
			done := make(chan struct{})
			go func() {
				defer close(done)
				m.runReadStack(ctx) // 运行协议栈读取循环
			}()

			select {
			case <-ctx.Done(): // 收到停止信号
				m.channelEpClose() // 关闭端点
				if m.tcpipStack != nil {
					m.tcpipStack.Destroy()
				}

				if m.tcpForwarder != nil {
					m.tcpForwarder.Close()
				}
				<-done // 等待任务完成
			case <-done: // 任务自然结束
				return
			}
		})

		//  读取 WinDivert 捕获的数据包
		m.tcm.AddTask(1, func(ctx context.Context) {
			done := make(chan struct{})
			go func() {
				defer close(done)
				m.runReadDivert(ctx) // 运行数据包捕获循环
			}()

			select {
			case <-ctx.Done():
				m.closeDev() // 关闭设备
				<-done       // 等待任务完成
			case <-done:
				return
			}
		})

		return m.exitChan, nil
	})

	return m
}

// Start 启动代理服务的各个组件
func (m *manager) Start() (<-chan error, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.start()
}

// Stop 停止所有服务组件
func (m *manager) Stop() {
	m.stop.Do(func() {
		m.mu.Lock()
		defer m.mu.Unlock()

		m.tcm.Stop()
		m.exitChanCloseFunc()

		m.closeDev()
		if m.channelEpClose != nil {
			m.channelEpClose() // 关闭端点
		}

		if m.tcpipStack != nil {
			m.tcpipStack.Destroy()
		}

		if m.tcpForwarder != nil {
			m.tcpForwarder.Close()
		}

		fmt.Println("透明rdp代理客户端退出")
	})
}
