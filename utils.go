package tcppool

import (
	"errors"
	"net"
)

//获取当前的conns
func (cp *ChannelPool) getConns() chan net.Conn {
	cp.mutex.Lock()
	conns := cp.conns
	cp.mutex.Unlock()
	return conns
}

//从池子里面获取一个资源
func (cp *ChannelPool) get() (net.Conn, error) {
	cp.mutex.Lock() //上锁
	if cp.conns == nil {
		cp.mutex.Unlock()
		return nil, errors.New("channelpool is closed")
	}
	cp.mutex.Unlock() //提前解锁皆因为channel本身自带读写锁
	select {
	case conn := <-cp.conns:
		if conn == nil {
			return nil, errors.New("channelpool is closed")
		}
		return conn, nil
	default:
		if cp.currentCapaciry >= cp.capacity {
			return nil, errors.New("Connection number to reach the upper limit")
		}
		//池子内获取不到连接，则初始化连接
		if conn, err := cp.factory(); err != nil {
			return nil, err
		} else {
			cp.mutex.Lock()
			cp.currentCapaciry += 1 //没投放一个连接到池子里面，把池塘的容量加1
			cp.mutex.Unlock()
			return conn, nil
		}
	}
}

//把拿到的资源回收到池子里面
func (cp *ChannelPool) put(conn net.Conn) error {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}
	if cp.conns == nil {
		return conn.Close()
	}
	select {
	case cp.conns <- conn:
		return nil
	default:
		return conn.Close()
	}
}
