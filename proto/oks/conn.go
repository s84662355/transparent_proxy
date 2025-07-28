package oks

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

func GetConn(ctx context.Context, s string, targetAddr string) (net.Conn, error) {
	// 1. 解析代理URL获取认证信息和服务器地址
	_, server, _, _, err := ParseURL(s)
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

	buf := make([]byte, 4)

	// 使用binary.BigEndian将整数写入字节切片（大端序）
	binary.BigEndian.PutUint32(buf, uint32(len(targetAddr)))

	buf = append(buf, []byte(targetAddr)...)

	// 发送消息到服务端
	_, err = conn.Write(buf)
	if err != nil {
		conn.Close()
		return nil, err
	}

	buf = make([]byte, 4)

	if _, err := io.ReadFull(conn, buf); err != nil {
		log.Printf("读取响应失败: %v", err)
		conn.Close()
		return nil, err
	}

	parsedNum := binary.BigEndian.Uint32(buf)

	buf = make([]byte, parsedNum)

	if _, err := io.ReadFull(conn, buf); err != nil {
		conn.Close()
		return nil, err
	}

	response := string(buf)

	log.Printf("读取响应: %s", response)

	if response != "ok" {
		conn.Close()
		return nil, fmt.Errorf(response)
	}

	return conn, nil
}
