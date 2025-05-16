package http

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/url"
)

// ParseURL 解析HTTP代理URL字符串
// s: HTTP代理URL字符串，格式为 [scheme://][user:password@]host:port[#domain]
// 返回值:
//
//	auth: Basic认证头信息(如"Basic dXNlcjpwYXNz")，无认证信息时为空
//	addr: 代理服务器地址(host:port格式)
//	domain: 目标域名(来自URL片段#部分)，默认为代理服务器host
//	scheme: 协议类型(当前仅支持http)
//	err: 解析过程中遇到的错误
func ParseURL(s string) (auth, addr, domain, scheme string, err error) {
	// 1. 使用标准库解析URL
	u, err := url.Parse(s)
	if err != nil {
		return // 返回URL解析错误
	}

	// 2. 处理认证信息
	if u.User != nil {
		username := u.User.Username()
		password, ok := u.User.Password()
		if !ok {
			err = errors.New("HTTP代理URL错误: 缺少密码")
			return
		}
		// 构造Basic认证头
		auth = fmt.Sprintf("%v:%v", username, password)
		auth = fmt.Sprintf("Basic %v", base64.StdEncoding.EncodeToString([]byte(auth)))
	}

	// 3. 处理协议和地址
	switch u.Scheme {
	case "http":
		host := u.Hostname()
		port := u.Port()
		if port == "" {
			port = "80" // 默认HTTP端口
		}
		addr = net.JoinHostPort(host, port) // 组合host和port

		// 4. 获取目标域名(来自URL片段)
		domain = u.Fragment
		if domain == "" {
			domain = host // 默认使用代理服务器host
		}

		scheme = "http" // 设置协议类型
	default:
		err = fmt.Errorf("不支持的协议类型: %v", u.Scheme)
		return
	}

	return
}
