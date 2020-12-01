package connection

import (
	"fmt"
	"net"
	"strconv"
	"syscall"
	"time"

	mdgoErr "github.com/aizsfgk/mdgo/net/error"
	"github.com/aizsfgk/mdgo/net/eventloop"
	"github.com/aizsfgk/mdgo/base/atomic"
	"github.com/aizsfgk/mdgo/net/event"
	"github.com/aizsfgk/mdgo/net/buffer"
	"github.com/aizsfgk/mdgo/net/callback"
)




type Connection struct {
	connFd int
	connected atomic.Bool
	inBuf *buffer.FixBuffer
	outBuf *buffer.FixBuffer
	cb callback.Callback

	peerAddr string
	eventLoop *eventloop.EventLoop
	activeTime  atomic.Int64
}


func New(fd int, loop *eventloop.EventLoop, sa syscall.Sockaddr, cb callback.Callback) (*Connection, error) {
	conn := &Connection{
		connFd:    fd,
		inBuf: buffer.NewFixBuffer(),
		outBuf: buffer.NewFixBuffer(),
		peerAddr:  sockAddrToString(sa),
		eventLoop: loop,
		cb: cb,
	}
	conn.connected.Set(true)
	return conn,nil
}


//func (conn *Connection) SetOnWriteComplete(msgCb callback.WriteCompleteCallback) {
//	conn.cb.OnWriteComplete = msgCb
//}
//
//func (conn *Connection) SetOnClose(msgCb callback.CloseCallback) {
//	conn.cb.OnClose = msgCb
//}


func (conn *Connection) Fd() int {
	return conn.connFd
}

func (conn *Connection) Close() error {
	if !conn.connected.Get() {
		return mdgoErr.ErrConnectionClosed
	}
	return conn.handleClose(conn.Fd())
}

func (conn *Connection) HandleEvent(eve event.Event, nowUnix int64) error {
	fmt.Println("Connection HandleEvent")
	fmt.Println("Connection-fd:", conn.Fd())
	fmt.Println("eve: ",eve)

	conn.activeTime.Swap(time.Now().Unix())

	var err error
	if eve & event.EventErr != 0 {
		err = conn.handleError(conn.Fd())
		if err != nil {
			fmt.Println("HandleError-err: ", err)
			return err
		}
	}

	if eve & event.EventRead != 0 {
		err = conn.handleRead(nowUnix)
		if err != nil {
			fmt.Println("handleRead-err: ", err)
			return err
		}
	}

	if eve & event.EventWrite != 0 {
		err = conn.handleWrite(conn.Fd())
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
func (conn *Connection) handleRead(nowUnix int64) error {

	// 等待几秒返回
	n, err := conn.inBuf.ReadFd(conn.Fd())
	if err.Temporary() { // 非阻塞会返回EAGAIN: resource temporarily unavailable
		return nil
	}
	fmt.Println("inbuf: ", conn.inBuf.RetrieveAllAsString())

	if n > 0 {
		// messageCallback回调使用
		conn.cb.OnMessage(nowUnix)

	} else if n == 0 {
		conn.handleClose(conn.Fd())
	} else {
		conn.handleError(conn.Fd())
	}


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