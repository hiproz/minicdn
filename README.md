## MiniCDN
极简内容分发系统

Use groupcache

Max cache size: 512M

```go
go get -u -v github.com/golang/groupcache
go build
```

## 架构

* M: Manager
	
	1. 负责维护Peer的列表,每个peer会去Manager同步这个列表。
	2. 所有的请求会先请求到manager, 然后由manager重定向到不同的peer

* P: Peer

	1. 提供文件的下载服务
	2. Peer之间会根据从manager拿到的peer列表，同步文件

Manager与Peer是一对多的关系

```
[M]
 |`------+--------+---......
 |       |        |
[S]     [S]      [S]  ....
```


### Run Manager
命令行启动

```shell
./minicdn -mirror http://localhost:5000 -addr :11000 -log cdn.log
```

* 对网站 `http://localhost:5000`进行镜像加速
* 监听11000端口
* 日志存储在cdn.log中

### Run Slave
命令行启动

```shell
./minicdn -upstream http://localhost:11000 -addr :8001
```

* 指定Server地址 `http://localhost:11000`
* 监听8001端口

### TODO
* token
* use a slave as a master
* request log
* cli args to specify cache size
