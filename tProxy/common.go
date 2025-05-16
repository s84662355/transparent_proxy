package tProxy

import (
	"sync"
	"time"
)

type ProxyJson struct {
	// 代理服务器地址 (格式: ip:port 或 domain:port)
	ProxyUrl string

	// 代理类型 (如: "http", "socks5", "trojan" 等)
	ProxyType string
	// Trojan代理配置 (当ProxyType为"trojan"时使用)
	TrojanProxy struct {
		// Trojan服务器地址 (格式: ip:port 或 domain:port)
		Server string

		// WebSocket路径 (用于WebSocket传输模式)
		Path string

		// Trojan连接密码
		Password string

		// 传输协议 (如: "ws", "tls" 等)
		Transport string

		// 域名 (用于TLS SNI和HTTP Host头)
		Domain string

		InsecureSkipVerify bool
	}
}

// item 存储缓存值和过期时间
type item struct {
	expiration int64 // UnixNano 时间戳
}

// TTLMap 带过期时间的线程安全Map
type TTLMap struct {
	mm    map[string]*item
	ttl   time.Duration // 默认过期时间
	mu    sync.Mutex    // 用于清理操作的锁
	start func()
	done  chan struct{}
}

// NewTTLMap 创建一个新的TTLMap
func NewTTLMap(defaultTTL time.Duration) *TTLMap {
	m := &TTLMap{
		ttl:  defaultTTL,
		done: make(chan struct{}),
	}

	m.start = sync.OnceFunc(func() {
		m.mm = map[string]*item{}

		go func() {
			ticker := time.NewTicker(3 * time.Second)
			defer close(m.done)
			defer ticker.Stop()
			for {
				select {
				case m.done <- struct{}{}:
					return
				case <-ticker.C:
					m.mu.Lock()
					if m.mm != nil {
						for k, v := range m.mm {
							if time.Now().UnixNano() >= v.expiration {
								delete(m.mm, k)
							}
						}
					}
					m.mu.Unlock()
				}
			}
		}()
	})

	return m
}

func (m *TTLMap) Start() {
	m.start()
}

func (m *TTLMap) Stop() {
	m.mu.Lock()
	m.mm = nil
	m.mu.Unlock()

	for range m.done {
	}
}

// Set 添加或更新键值对
func (m *TTLMap) Set(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mm != nil {
		m.mm[key] = &item{
			expiration: time.Now().Add(m.ttl).UnixNano(),
		}
	}
}

// Get 获取值并可选续期
func (m *TTLMap) Get(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mm == nil {
		return false
	}

	if it, ok := m.mm[key]; !ok {
		return false
	} else {
		it.expiration = time.Now().Add(m.ttl).UnixNano()
	}

	return true
}

// Delete 删除键
func (m *TTLMap) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mm == nil {
		return
	}

	delete(m.mm, key)
}
