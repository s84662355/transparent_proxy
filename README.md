# transparent_proxy

## 编译
```shell
go build  -o bin/proxy.exe 
```


## 代码格式化
```shell
 gofumpt -l -w .
```

## 启动命令
```shell
./proxy.exe  --config_path=conf
```


## 配置文件 conf

```shell
{
	"ProxyUrl":"socks5://test985:test985@8.219.163.116:7777",
	"ProxyType":"socks"
}
 
```



