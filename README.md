# vnc-relay-server

## 编译
```shell
go build  -o bin/proxy.exe 
```


## 代码格式化
```shell
 gofumpt -l -w .
```

```shell
./proxy.exe  --config_path=conf
```


## conf

```shell
{
	"ProxyUrl":"socks5://test985:test985@8.219.163.116:7777",
	"ProxyType":"socks"
}
 
```



 