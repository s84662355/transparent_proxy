package config

type confData struct {
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
