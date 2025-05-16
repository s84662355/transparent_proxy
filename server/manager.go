package server

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"transparent/config"

	"transparent/log"
	"transparent/tProxy"
	"transparent/utils/taskConsumerManager"
)

// 使用 sync.OnceValue 确保 manager 只被初始化一次（线程安全）
var newManager = sync.OnceValue(func() *manager {
	m := &manager{
		tcm: taskConsumerManager.New(), // 任务消费者管理器
	}

	return m
})

// manager 结构体管理整个代理服务的核心组件
type manager struct {
	tcm *taskConsumerManager.Manager // 任务调度管理器
}

// Start 启动代理服务的各个组件
func (m *manager) Start() error {
	proxyJson := &tProxy.ProxyJson{}
	proxyJson.ProxyUrl = config.GetConf().ProxyUrl
	proxyJson.ProxyType = config.GetConf().ProxyType
	proxyJson.TrojanProxy = config.GetConf().TrojanProxy

	m.tcm.AddTask(1, func(ctx context.Context) {
		t := tProxy.NewManager(proxyJson)
		eCh, err := t.Start()
		if err != nil {
			log.Error("创建代理对象失败", zap.Error(err))
			return
		}
		defer t.Stop()
		select {
		case <-ctx.Done():
		case <-eCh:
		}
	})

	return nil
}

// Stop 停止所有服务组件
func (m *manager) Stop() {
	m.tcm.Stop() // 停止任务消费者管理器，会触发所有任务的优雅关闭
}
