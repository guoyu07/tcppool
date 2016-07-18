package tcppool

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"
)

var (
	initialCap = 5
	maxCap     = 30
	network    = "tcp"
	address    = "127.0.0.1:7777"
	factory    = func() (net.Conn, error) { return net.Dial(network, address) }
)

func init() {
	go simpleTCPServer()
	time.Sleep(time.Millisecond * 300) // wait until tcp server has been settled
	rand.Seed(time.Now().UTC().UnixNano())
}

//go test -run='TestNew'
//测试整个package是否有语法错误
func TestNew(t *testing.T) {
	_, err := NewChannelPool(initialCap, maxCap, factory)
	if err != nil {
		t.Errorf("New error: %s", err)
	}
}

//go test -run='Test_Channelpool_GetType'
//测试get到的连接是否是net.Conn
func Test_Channelpool_GetType(t *testing.T) {
	pool, err := NewChannelPool(initialCap, maxCap, factory)
	if err != nil {
		t.Errorf("New error: %s", err)
	}
	defer pool.Close()
	conn, err := pool.Get()
	if err != nil {
		t.Errorf("Get error: %s", err)
	}
	_, ok := conn.(net.Conn)
	if !ok {
		t.Errorf("Conn is not of type net.Conn")
	}
}

//go test -run='Test_Channelpool_Get'
func Test_Channelpool_Get(t *testing.T) {
	pool, err := NewChannelPool(initialCap, maxCap, factory)
	if err != nil {
		t.Errorf("New error: %s", err)
	}
	defer pool.Close()
	_, err = pool.Get()
	if err != nil {
		t.Errorf("Get error: %s", err)
	}
	//当从连接池内获取一个连接时，池子内的连接减1
	if pool.Len() != initialCap-1 {
		t.Errorf("Get error. Expecting %d, got %d", (initialCap - 1), pool.Len())
	}
	fmt.Println(pool.Len())
	var wg sync.WaitGroup
	for i := 0; i < (initialCap - 1); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := pool.Get()
			if err != nil {
				t.Errorf("Get error: %s", err)
			}
		}()
	}
	wg.Wait()
	fmt.Println(pool.Len())
	if pool.Len() != 0 {
		t.Errorf("Get error. Expecting %d, got %d", (initialCap - 1), pool.Len())
	}
	_, err = pool.Get() //当池内的连接为空，且不超过最大连接时，还能获取到
	if err != nil {
		t.Errorf("Get error: %s", err)
	}
	// for i := 0; i < 30; i++ { //超过最大连接池报错
	// 	_, err = pool.Get() //当池内的连接为空，且不超过最大连接时，还能获取到
	// 	if err != nil {
	// 		t.Errorf("Get error: %s", err)
	// 	}
	// }
	fmt.Println(pool.Len())
}

//go test -run='Test_Channelpool_Put'
func Test_Channelpool_Put(t *testing.T) {
	pool, err := NewChannelPool(initialCap, maxCap, factory)
	if err != nil {
		t.Errorf("New error: %s", err)
	}
	defer pool.Close()
	conn, err := pool.Get()
	if err != nil {
		t.Errorf("Get error: %s", err)
	}
	pool.Put(conn)
	fmt.Println(pool.Len())
	if pool.Len() != initialCap {
		t.Errorf("Get error. Expecting %d, got %d", (initialCap - 1), pool.Len())
	}
}
func Test_Channelpool_Close(t *testing.T) {
	pool, err := NewChannelPool(initialCap, maxCap, factory)
	if err != nil {
		t.Errorf("New error: %s", err)
	}
	pool.Close()
	c := pool.(*ChannelPool)

	if c.conns != nil {
		t.Errorf("Close error, conns channel should be nil")
	}

	if c.factory != nil {
		t.Errorf("Close error, factory should be nil")
	}
	_, err = pool.Get()
	if err == nil {
		t.Errorf("Close error, get conn should return an error")
	}
	if pool.Len() != 0 {
		t.Errorf("Close error used capacity. Expecting 0, got %d", pool.Len())
	}
}

func TestPoolConcurrent(t *testing.T) {
	pool, err := NewChannelPool(initialCap, maxCap, factory)
	if err != nil {
		t.Errorf("New error: %s", err)
	}
	pipe := make(chan net.Conn, 0)
	defer pool.Close()
	for i := 0; i < maxCap; i++ {
		go func() {
			conn, _ := pool.Get()

			pipe <- conn
		}()

		go func() {
			conn := <-pipe
			if conn == nil {
				return
			}
			conn.Close()
		}()
	}
}
func TestPoolWriteRead(t *testing.T) {
	pool, err := NewChannelPool(initialCap, maxCap, factory)
	if err != nil {
		t.Errorf("New error: %s", err)
	}
	defer pool.Close()

	conn, _ := pool.Get()

	msg := "hello"
	_, err = conn.Write([]byte(msg))
	if err != nil {
		t.Error(err)
	}
}

//开启一个简单的tcp server
func simpleTCPServer() {
	l, err := net.Listen(network, address)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			buffer := make([]byte, 256)
			conn.Read(buffer)

		}()
	}
}
