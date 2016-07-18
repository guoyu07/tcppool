package tcppool

import (
	"net"
)

/*
	tcppool interface是对这个连接池的一个描述。也暴露了对外的三个接口。
	一个优秀的连接池要能实现对池子的大小控制，线程取用安全，简单等。
*/
type TcpPool interface {
	//从连接池内获取到一个连接
	Get() (net.Conn, error)
	//把获取到的连接重新放回池子里面
	Close()
	//把资源放回池子里面
	Put(conn net.Conn) error
	//查看当前连接池的大小
	Len() int
}
