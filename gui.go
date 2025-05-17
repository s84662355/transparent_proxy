//go:build gui
// +build gui

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"transparent/log"
	"transparent/server"

	"go.uber.org/zap"
)

func init() {
	// 获取可执行文件的绝对路径
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	exeDir := filepath.Dir(exePath)
	fmt.Println(filepath.Join(exeDir, "log"))
	log.Init(filepath.Join(exeDir, "log"))
	go gohttp()
}

// main 程序主入口

func main() {
	// 1. 启动服务器
	err := server.Start()
	if err != nil {
		// 启动失败记录致命错误并退出
		log.Fatal("服务器启动失败",
			zap.String("error", err.Error()))
		return
	}
	server.Stop() // 停止服务器
}
