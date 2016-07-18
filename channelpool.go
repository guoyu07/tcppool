package tcppool

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

type ChannelPool struct {
	conns           chan net.Conn //定义一个conn类型的channel
	mutex           sync.Mutex    //锁
	factory         Factory       //Factory类型
	capacity        int
	currentCapaciry int
}

//定义一个factory的方法
//func() (net.Conn, error) { return net.Dial("tcp", "127.0.0.1:4000") }
type Factory func() (net.Conn, error)

//初始化channel Pool
func NewChannelPool(initialCap, maxCap int, factory Factory) (TcpPool, error) {
	if initialCap < 0 || maxCap <= 0 || initialCap > maxCap {
		return nil, errors.New("Failed to initialize connection pool capacity") //初始化连接池容量失败
	}
	cp := new(ChannelPool)                 //初始化channelpool
	cp.conns = make(chan net.Conn, maxCap) //初始化conns这个channel；上面代码保证了maxCap不能小于等于0
	cp.factory = factory                   //把外部传近来的方法赋值给factory
	cp.capacity = maxCap                   //初始化连接池的最大容量标示
	cp.currentCapaciry = 0                 //初始化当前连接池的容量
	for i := 0; i < initialCap; i++ {      //根据池子初始化的大小对池子进行初始化
		if conn, err := cp.factory(); err != nil {
			cp.Close() //对池子进行清理
			return nil, fmt.Errorf("The factory is unable to fill connection pool. err : %s", err)
		} else {
			cp.conns <- conn //给池子放点育苗
			cp.mutex.Lock()
			cp.currentCapaciry += 1 //没投放一个连接到池子里面，把池塘的容量加1
			cp.mutex.Unlock()
		}
	}
	return cp, nil
}

//从池子里面获取资源
func (cp *ChannelPool) Get() (net.Conn, error) {
	conn, err := cp.get()
	return conn, err
}

//把资源回收到池子里面
func (cp *ChannelPool) Put(conn net.Conn) error {
	return cp.put(conn)
}

//获取当前连接池的大小
func (cp *ChannelPool) Len() int {
	return len(cp.getConns())
}

//关闭释放掉池子内的所有资源
func (cp *ChannelPool) Close() {
	if cp.conns == nil { //说明之前已经关闭掉了
		return
	}
	cp.mutex.Lock() //上锁
	close(cp.conns) //关闭池子
	for conn := range cp.conns {
		conn.Close() //关闭池子内存在的所有TCP连接
	}
	cp.factory, cp.conns, cp.capacity, cp.currentCapaciry = nil, nil, 0, 0
	cp.mutex.Unlock() //解锁
}
