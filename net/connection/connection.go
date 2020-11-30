package connection

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"

	mdgoErr "github.com/aizsfgk/mdgo/net/error"
	"github.com/aizsfgk/mdgo/net/eventloop"
	"github.com/aizsfgk/mdgo/base/atomic"
	"github.com/aizsfgk/mdgo/net/event"
)


type Connection struct {
	connFd int
	connected atomic.Bool
	inBuf bytes.Buffer
	outBuf bytes.Buffer

	peerAddr string
	eventLoop *eventloop.EventLoop
	activeTime  atomic.Int64
}


func New(fd int, loop *eventloop.EventLoop, sa syscall.Sockaddr) (*Connection, error) {
	conn := &Connection{
		connFd:    fd,
		inBuf: bytes.Buffer{},
		outBuf: bytes.Buffer{},
		peerAddr:  sockAddrToString(sa),
		eventLoop: loop,
	}
	conn.connected.Set(true)
	return conn,nil
}


func (conn *Connection) Fd() int {
	return conn.connFd
}

func (conn *Connection) Close() error {
	if !conn.connected.Get() {
		return mdgoErr.ErrConnectionClosed
	}
	return conn.handleClose(conn.Fd())
}

func (conn *Connection) HandleEvent(fd int, eve event.Event) error {
	fmt.Println("Connection HandleEvent")
	fmt.Println("Connection-fd:", fd)
	fmt.Println("eve: ",eve)

	conn.activeTime.Swap(time.Now().Unix())

	var err error
	if eve & event.EventErr != 0 {
		err = conn.handleError(fd)
		if err != nil {
			fmt.Println("HandleError-err: ", err)
			return err
		}
	}

	if eve & event.EventRead != 0 {
		err = conn.handleRead(fd)
		if err != nil {
			fmt.Println("handleRead-err: ", err)
			return err
		}
	}

	if eve & event.EventWrite != 0 {
		err = conn.handleWrite(fd)
		if err != nil {
			fmt.Println("handleWrite-err: ", err)
			return err
		}
	}

	return nil
}

//
/**
 * 处理读
 *   1. 读就绪，如果不处理，（水平触发下）会一直通知；
 *   因为此时：接收缓冲区中一直有数据，水平触发下，需要一直通知
 */
func (conn *Connection) handleRead(fd int) error {
	fmt.Println("handleRead: fd:", fd)

	// 读取数据

	// 返回0(EOF), 则关闭连接

	// 返回-1， 表示有错误发生

	return nil
}

// 2. 处理写

func (conn *Connection) handleWrite(fd int) error {
	fmt.Println("handleWrite: fd:", fd)

	return nil
}

// 3. 处理关闭

func (conn *Connection) handleClose(fd int) error {
	fmt.Println("handleClose: fd:", fd)

	return syscall.Close(fd)
}

// 4. 处理错误

func (conn *Connection) handleError(fd int) error {
	fmt.Println("handleError: fd:", fd)
	return nil
}



func sockAddrToString(sa syscall.Sockaddr) string {
	switch sa := (sa).(type) {
	case *syscall.SockaddrInet4:
		return net.JoinHostPort(net.IP(sa.Addr[:]).String(), strconv.Itoa(sa.Port))
	case *syscall.SockaddrInet6:
		return net.JoinHostPort(net.IP(sa.Addr[:]).String(), strconv.Itoa(sa.Port))
	default:
		return fmt.Sprintf("(unknow - %T)", sa)
	}
}