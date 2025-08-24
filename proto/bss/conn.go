package bss

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"slices"
	"time"
)

// 全局加密密钥（实际使用中应通过安全方式分发，长度必须为16/24/32字节）
var encryptionKey = []byte("0123456789abcdef") // 16字节，AES-128

// 加密数据：使用AES-GCM算法
func encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	// GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 生成随机非ce（12字节推荐值）
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// 加密并添加认证标签
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// 解密数据：对应AES-GCM加密
func decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 分离非ce和密文
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("密文太短")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// 发送加密数据：先发送长度（4字节大端序），再发送加密内容
func sendEncrypted(conn net.Conn, data []byte) (int, error) {
	// 加密数据
	encrypted, err := encrypt(data)
	if err != nil {
		return 0, err
	}

	// 先发送长度（4字节）
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(encrypted)))
	if _, err := conn.Write(lengthBuf); err != nil {
		return 0, err
	}

	// 再发送加密内容
	return conn.Write(encrypted)
}

// 接收加密数据：先读长度，再读内容并解密
func receiveEncrypted(conn net.Conn) ([]byte, error) {
	// 读取长度（4字节）
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lengthBuf); err != nil {
		return nil, err
	}
	dataLen := binary.BigEndian.Uint32(lengthBuf)

	// 读取加密内容
	encrypted := make([]byte, dataLen)
	if _, err := io.ReadFull(conn, encrypted); err != nil {
		return nil, err
	}

	// 解密
	return decrypt(encrypted)
}

// Base64Encode 对二进制数据进行Base64编码
func Base64Encode(data []byte) string {
	// 使用标准Base64编码（兼容RFC 4648）
	return base64.StdEncoding.EncodeToString(data)
}

// Base64Decode 对Base64编码的字符串进行解码
func Base64Decode(encoded string) ([]byte, error) {
	// 解码Base64字符串，返回原始二进制数据
	return base64.StdEncoding.DecodeString(encoded)
}

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

	targetAddr = Base64Encode([]byte(targetAddr))

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
	if response != "ok" {
		conn.Close()
		return nil, fmt.Errorf(response)
	}

	c := &Conn{
		conn: conn,
	}

	return c, nil
}

type Conn struct {
	conn net.Conn
	data []byte
}

func (c *Conn) Read(b []byte) (n int, err error) {
	if len(c.data) > 0 {
		if len(c.data) > len(b) {
			n = copy(b, c.data[:len(b)])
			c.data = c.data[:n]
			return
		} else {
			n = copy(b, c.data)
			c.data = nil
			return
		}
	}

	d, err := receiveEncrypted(c.conn)
	if err != nil {
		return 0, err
	}
	c.data = d

	if len(c.data) > 0 {
		if len(c.data) > len(b) {
			n = copy(b, c.data[:len(b)])
			c.data = c.data[:n]
			return
		} else {
			n = copy(b, c.data)
			c.data = nil
			return
		}
	}

	return
}

func (c *Conn) Write(b []byte) (n int, err error) {
	for v := range slices.Chunk(b, 32*1024) {
		if _, errt := sendEncrypted(c.conn, v); errt != nil {
			fmt.Println(errt)
			return 0, errt
		}
	}

	return len(b), nil
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
