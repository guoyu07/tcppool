# Pool [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/gopkg.in/fatih/pool.v2) [![Build Status](http://img.shields.io/travis/fatih/pool.svg?style=flat-square)](https://travis-ci.org/fatih/pool)

TcpPool is a thread safe connection pool for net.Conn interface. It can be used to manage and reuse connections.

## Install and Usage

Install the package with:

```bash
go get github.com/bibinbin/tcppool
```

Import it with:

```go
import "github.com/bibinbin/tcppool"
```
and use `pool` as the package name inside the code.

```go
import (
  pool "github.com/bibinbin/tcppool"
)
```

## Example
```go

factory    := func() (net.Conn, error) { return net.Dial("tcp", "127.0.0.1:5000") }
pool, err := NewChannelPool(5, 30, factory)   //初始化，tcppool
if err != nil {
	log.Println(err)
}
defer pool.Close()  //关闭tcppool
conn, err := pool.Get() //从连接池内获取一个资源
if err != nil {
	log.Println(err)
}
pool.Put(conn) //把资源重现填充到池子内
pool.Len() //获取当前连接池的长度

```

