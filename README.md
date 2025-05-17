# transparent_proxy

## 编译gui版本
```shell
go build -tags "gui" -o bin/gui.exe 
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
	"ProxyUrl":"socks5://test985:test985@8.219.163.116:7777",
	"ProxyType":"socks"
}
 
```

## gui版本截图
<img src="assets/gui.png" alt="界面截图">
<img src="assets/gui1.png" alt="界面截图">



