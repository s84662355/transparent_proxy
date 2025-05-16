package http

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
)

// GetConn 通过HTTP CONNECT代理建立到目标地址的隧道连接
// s: 代理服务器URL，格式为 [http://][user:password@]host:port
// targetAddr: 要通过代理连接的目标地址，格式为 "host:port"
// 返回值: 已建立的隧道连接和可能的错误
func GetConn(ctx context.Context, s string, targetAddr string) (net.Conn, error) {
	// 1. 解析代理URL获取认证信息和服务器地址
	auth, server, _, _, err := ParseURL(s)
	if err != nil {
		return nil, fmt.Errorf("解析代理URL失败: %w", err)
	}

	// 2. 创建TCP拨号器并连接到代理服务器
	var Dialer net.Dialer
	conn, err := Dialer.DialContext(ctx, "tcp", server)
	if err != nil {
		return nil, fmt.Errorf("连接代理服务器失败: %w", err)
	}

	done := make(chan struct{})
	defer func() {
		for range done {
		}
	}()

	///通过ctx控制http代理的握手超时
	go func() {
		defer close(done)
		select {
		case <-ctx.Done():
			conn.Close()
		case done <- struct{}{}:
		}
	}()

	// 3. 准备CONNECT请求
	req, err := http.NewRequest(http.MethodConnect, "", nil) // 空路径表示CONNECT隧道
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("创建CONNECT请求失败: %w", err)
	}

	// 4. 设置请求头
	req.Host = targetAddr // 设置要连接的目标地址
	if auth != "" {
		req.Header.Add("Proxy-Authorization", auth) // 添加代理认证头
	}

	// 5. 发送CONNECT请求到代理服务器
	if err = req.Write(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("发送CONNECT请求失败: %w", err)
	}

	// 6. 读取代理服务器响应
	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, req)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("读取代理响应失败: %w", err)
	}
	defer resp.Body.Close()

	// 7. 检查响应状态码(需要200表示成功)
	if resp.StatusCode != http.StatusOK {
		conn.Close()
		return nil, fmt.Errorf("代理服务器返回错误状态码: %d", resp.StatusCode)
	}

	// 8. 处理可能的缓冲数据(防止粘包)
	if n := reader.Buffered(); n > 0 {
		b := make([]byte, n)
		if _, err = io.ReadFull(conn, b); err != nil {
			conn.Close()
			return nil, fmt.Errorf("读取缓冲数据失败: %w", err)
		}
	}

	// 9. 返回已建立的隧道连接
	return conn, nil
}
