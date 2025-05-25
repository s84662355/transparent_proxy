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

	"golang.org/x/sys/windows"

	"transparent/utils/windowsMutex"
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
	AppName := "transparent_proxy"

	{
		//###############不会重复打开应用
		mutex, err := windowsMutex.WindowsMutex(AppName)
		if err != nil {
			log.Error(AppName+" WindowsMutex  mutex error ", zap.Error(err))
			return
		}

		defer func() {
			windows.ReleaseMutex(windows.Handle(mutex))
			windows.CloseHandle(windows.Handle(mutex))
		}()

		event, err := windows.WaitForSingleObject(mutex, windows.INFINITE)
		if err != nil {
			log.Error(AppName+" wait for mutex error", zap.Error(err))
			return
		}

		switch event {
		case windows.WAIT_OBJECT_0, windows.WAIT_ABANDONED:
		default:
			log.Error(AppName+" wait for mutex event id error: ", zap.Any("event", event))
			return
		}
		//########################
	}

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
