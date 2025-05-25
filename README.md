# transparent_proxy

透明全局代理，暂时支持socks和http代理

## 编译前准备
```shell

# 安装 rsrc 工具（确保 GOPATH/bin 在 PATH 环境变量中）
go install github.com/akavel/rsrc@latest

# 生成 .syso 文件
rsrc -manifest app.manifest -o app.syso

```


## 编译gui版本
```shell
go build -tags "gui" -ldflags="-H windowsgui"    -o bin/gui.exe 
```

## 编译console版本
```shell
go build -tags "console" -o bin/console.exe 
```


## 代码格式化
```shell
 gofumpt -l -w .
```

## 启动命令 控制台版本
```shell
./console.exe  --config_path=conf
```


## 配置文件 conf

```shell
{
	"ProxyUrl":"socks5://t766785:t765@81.221.16.16:19577",
	"ProxyType":"socks"
}
 
```

## gui版本截图
<img src="assets/gui.png" alt="界面截图">
<img src="assets/gui1.png" alt="界面截图带代理">



