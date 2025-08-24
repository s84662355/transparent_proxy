package gui

import (
	//"context"
	"fmt"
	"net/url"
	"sync"

	//	"go.uber.org/zap"

	//"transparent/config"

	//	"transparent/log"
	"transparent/tProxy"
	"transparent/utils/taskConsumerManager"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// 使用 sync.OnceValue 确保 manager 只被初始化一次（线程安全）
var NewManager = sync.OnceValue(func() *manager {
	m := &manager{
		tcm: taskConsumerManager.New(), // 任务消费者管理器
	}

	return m
})

// manager 结构体管理整个代理服务的核心组件
type manager struct {
	tcm       *taskConsumerManager.Manager // 任务调度管理器
	proxtT    tProxy.Manager
	proxyMu   sync.RWMutex
	a         fyne.App
	startBut  *widget.Button
	cancelBut *widget.Button
}

// Start 启动代理服务的各个组件
func (m *manager) Start() error {
	m.a = app.New()

	w := m.a.NewWindow("透明代理")

	// 固定窗口大小，禁止调整
	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(400, 350))

	input := widget.NewEntry()
	input.SetPlaceHolder("请输入文本内容...")
	input.Resize(fyne.NewSize(400, 100))

	m.startBut = widget.NewButton("启动", func() {
		text := input.Text

		u, err := url.Parse(text)
		if err != nil {
			fmt.Println("解析错误:", err)
			dialog.ShowError(err, w)
			return
		}

		proxyJson := &tProxy.ProxyJson{}

		switch u.Scheme {
		case "http":
			proxyJson.ProxyType = "http"
			proxyJson.ProxyUrl = text
		case "socks":
			proxyJson.ProxyUrl = text
			proxyJson.ProxyType = "socks"
		case "socks5":
			proxyJson.ProxyUrl = text
			proxyJson.ProxyType = "socks"
		case "oks":
			proxyJson.ProxyUrl = text
			proxyJson.ProxyType = "oks"
		case "bss":
			proxyJson.ProxyUrl = text
			proxyJson.ProxyType = "bss"

		default:
			dialog.ShowError(fmt.Errorf("不支持类型%s", u.Scheme), w)
			return
		}

		m.proxyMu.Lock()
		if m.proxtT != nil {
			m.proxtT.Stop()
		}
		proxtT := tProxy.NewManager(proxyJson)
		m.proxtT = proxtT
		m.proxyMu.Unlock()

		eCh, err := proxtT.Start()
		if err != nil {
			proxtT.Stop()
			dialog.ShowError(err, w)
		} else {
			m.startBut.Disable()
			go func() {
				<-eCh
				proxtT.Stop()
				fmt.Println("代理关闭")
			}()
		}
	})

	m.cancelBut = widget.NewButton("取消", func() {
		m.proxyMu.Lock()
		if m.proxtT != nil {
			m.proxtT.Stop()
		}
		m.proxtT = nil
		m.proxyMu.Unlock()
		m.startBut.Enable()
		fmt.Println("代理关闭")
	})

	// 创建按钮容器，水平排列并居中
	buttonContainer := container.NewHBox(
		m.startBut,
		m.cancelBut,
	)

	// 使用Border布局，将按钮容器居中
	buttonWrapper := container.NewBorder(
		nil, nil, // 上下无内容
		nil, nil, // 左右无内容
		buttonContainer,
	)

	// 主容器，垂直排列输入框和按钮行
	content := container.NewVBox(
		input,
		buttonWrapper,
	)

	w.SetContent(content)
	w.ShowAndRun()
	m.proxyMu.RLock()
	proxtT := m.proxtT
	m.proxyMu.RUnlock()
	if proxtT != nil {
		proxtT.Stop()
	}

	return nil
}

// Stop 停止所有服务组件
func (m *manager) Stop() {
	m.proxyMu.RLock()
	proxtT := m.proxtT
	m.proxyMu.RUnlock()
	if proxtT != nil {
		proxtT.Stop()
	}
}
