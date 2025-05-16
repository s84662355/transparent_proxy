package taskConsumerManager

import (
	"context"
	"sync"
)

type Manager struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func New() *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (m *Manager) Context() context.Context {
	return m.ctx
}

type taskFunc func(context.Context)

func (m *Manager) AddTask(count int, fc taskFunc) {
	select {
	case <-m.ctx.Done():
		return
	default:
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		tChan := make(chan struct{}, count)
		defer close(tChan)
		wg := sync.WaitGroup{}
		defer wg.Wait()
		for {
			select {
			case <-m.ctx.Done():
				return
			case tChan <- struct{}{}:
				select {
				case <-m.ctx.Done():
					return
				default:
				}
				wg.Add(1)
				go func() {
					defer func() {
						wg.Done()
						<-tChan
					}()
					fc(m.ctx)
				}()
			}
		}
	}()
}

func (m *Manager) Stop() {
	m.cancel()
	m.wg.Wait()
}
