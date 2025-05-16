package socks

import (
	"errors"
	"net/url"

	"golang.org/x/net/proxy"
)

// ParseURL 解析SOCKS5代理URL字符串
// s: SOCKS5代理URL字符串，格式为 [scheme://][user:password@]host:port
// 返回值:
//
//	auth: 认证信息结构体指针，如果URL中不包含用户信息则为nil
//	server: 代理服务器地址(host:port格式)
//	err: 解析过程中遇到的错误
func ParseURL(s string) (auth *proxy.Auth, server string, err error) {
	// 1. 使用标准库解析URL
	u, er := url.Parse(s)
	if er != nil {
		err = er // 返回URL解析错误
		return
	}

	// 2. 提取代理服务器地址(包含端口)
	server = u.Host

	// 3. 检查URL中是否包含用户认证信息
	if u.User == nil {
		return // 无认证信息，直接返回
	}

	// 4. 提取用户名和密码
	username := u.User.Username()
	password, ok := u.User.Password()
	if !ok {
		// 有用户名但没有密码的情况
		err = errors.New("socks url error: no password")
		return
	}

	// 5. 构造认证信息结构体
	auth = &proxy.Auth{
		User:     username,
		Password: password,
	}
	return
}
