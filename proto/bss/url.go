package bss

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/url"
)

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
	case "bss":
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

		scheme = "bss" // 设置协议类型
	default:
		err = fmt.Errorf("不支持的协议类型: %v", u.Scheme)
		return
	}

	return
}
