package trojan

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/Dreamacro/clash/transport/socks5"
)

func GetConn(ctx context.Context, serverAddr string, password string, targetAddr string, InsecureSkipVerify bool) (net.Conn, error) {
	// 解析目标地址为 Socks5 地址格式
	socks5Addr := socks5.ParseAddr(targetAddr)

	// 2. 创建TCP拨号器并连接到代理服务器
	var Dialer net.Dialer
	conn, err := Dialer.DialContext(ctx, "tcp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("连接代理服务器失败: %w", err)
	}

	// 创建 TLS 连接
	tlsConfig := &tls.Config{
		InsecureSkipVerify: InsecureSkipVerify, // 注意：生产环境中不要使用此选项
	}

	done := make(chan struct{})
	defer func() {
		for range done {
		}
	}()
	go func() {
		defer close(done)
		select {
		case <-ctx.Done():
			conn.Close()
		case done <- struct{}{}:
		}
	}()

	tlsConn := tls.Client(conn, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("Failed to perform TLS handshake: %w", err)
	}

	// 计算密码的 SHA224 哈希值
	hexPassword := hexSha224([]byte(password))

	// 写入 Trojan 头部
	if err := writeHeader(tlsConn, hexPassword, CommandTCP, socks5Addr); err != nil {
		conn.Close()
		return nil, fmt.Errorf("Failed to write Trojan header:  %w", err)
	}

	return tlsConn, nil
}
