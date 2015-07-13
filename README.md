## MiniCDN
极简内容分发系统

Use groupcache

Max cache size: 512M

```go
go get -u -v github.com/golang/groupcache
go build
```

### Run Server
* 对网站 `http://localhost:5000`进行镜像加速
* 监听11000端口
* 日志存储在cdn.log中

```shell
./minicdn -mirror http://localhost:5000 -addr :11000 -log cdn.log
```

### Run Slave
* 指定Server地址 `http://localhost:11000`
* 监听8001端口

```shell
./minicdn -upstream http://localhost:11000 -addr :8001
```

### TODO
* token
* use a slave as a master
* request log
* cli args to specify cache size
