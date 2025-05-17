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
	"ProxyUrl":"socks5://tes85:te985@8.21.16.16:6577",
	"ProxyType":"socks"
}
 
```



