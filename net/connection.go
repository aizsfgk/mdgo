package net

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/aizsfgk/mdgo/base/atomic"
	"github.com/aizsfgk/mdgo/net/buffer"
	mdgoErr "github.com/aizsfgk/mdgo/net/errors"
	"github.com/aizsfgk/mdgo/net/event"
)

// 定义回调接口
type Callback interface {
	OnMessage(*Connection, int64)
	OnClose()
	OnWriteComplete()
}

// 定义连接
type Connection struct {
	connFd     int               // acceptFd
	connected  atomic.Bool       // state[connected or not]
	InBuf      *buffer.FixBuffer // input buffer
	OutBuf     *buffer.FixBuffer // output buffer
	cb         Callback          // cb
	peerAddr   string            // remote addr
	eventLoop  *EventLoop        // work sub eventLoop
	activeTime atomic.Int64      // last active time
}

// 新建连接
func NewConnection(fd int, loop *EventLoop, sa syscall.Sockaddr, cb Callback) (*Connection, error) {
	conn := &Connection{
		connFd:    fd,
		InBuf:     buffer.NewFixBuffer(),
		OutBuf:    buffer.NewFixBuffer(),
		peerAddr:  sockAddrToString(sa),
		eventLoop: loop,
		cb:        cb,
	}
	conn.connected.Set(true)
	return conn, nil
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

//
// export api func
//

// close
func (conn *Connection) Close() error {
	if !conn.connected.Get() {
		return mdgoErr.ErrConnectionClosed
	}
	return conn.handleClose()
}

// send
func (conn *Connection) SendString(out string) error {
	return conn.SendByte([]byte(out))
}

func (conn *Connection) SendByte(out []byte) error {
	if !conn.connected.Get() {
		return mdgoErr.ErrConnectionClosed
	}

	_ = conn.Send(out)
	return nil
}


// 直接写回
// 如果输出缓冲不是空
// TODO 或者正在关注写事件，则追加数据
func (conn *Connection) Send(out []byte) (rerr error) {
	if conn.OutBuf.ReadableBytes() > 0 {
		conn.OutBuf.Append(out)
	} else {
		n, err := syscall.Write(conn.Fd(), out)
		if err != nil {
			// EAGAIN 说明没有数据空间，可以写入
			// n个字节追加到缓冲区
			/*
			普通做法：
			当需要向socket写数据时，将该socket加入到epoll等待可写事件。接收到socket可写事件后，调用write()或send()发送数据，当数据全部写完后， 将socket描述符移出epoll列表，这种做法需要反复添加和删除。

			改进做法:
			向socket写数据时直接调用send()发送，当send()返回错误码EAGAIN，才将socket加入到epoll，等待可写事件后再发送数据，全部数据发送完毕，再移出epoll模型，改进的做法相当于认为socket在大部分时候是可写的，不能写了再让epoll帮忙监控。上面两种做法是对LT模式下write事件频繁通知的修复，本质上ET模式就可以直接搞定，并不需要用户层程序的补丁操作。
			*/
			if err != syscall.EAGAIN {
				rerr = conn.handleClose()
				return
			}
			fmt.Println("write fd err: ", err)
		}

		// some condition, append bytes to out buffer
		if n < len(out) {
			fmt.Println("write fd, n: ", n)
			if n == 0 {
				conn.OutBuf.Append(out)
			} else {
				conn.OutBuf.Append(out[n:])
			}
		}

		// if out buffer has readable byte, enable fd write event
		if conn.OutBuf.ReadableBytes() > 0 {
			return conn.eventLoop.EnableReadWrite(conn.Fd())
		}
	}
	return nil
}

// ********* handle Event *********** //
func (conn *Connection) HandleEvent(eve event.Event, nowUnix int64) error {
	conn.activeTime.Swap(time.Now().Unix())

	var err error
	if eve&event.EventError != 0 {
		conn.handleError(conn.Fd())
		// TODO close conn
	}

	if eve&event.EventRead != 0 {
		err = conn.handleRead(nowUnix)
		if err != nil {
			fmt.Println("handleRead-err: ", err)
			return err
		}
	}

	if eve&event.EventWrite != 0 {
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
	n, err := conn.InBuf.ReadFd(conn.Fd())
	if err.Temporary() { // 非阻塞会返回EAGAIN: resource temporarily unavailable
		return nil
	}

	if n > 0 {
		// cb 2
		// messageCallback回调使用
		conn.cb.OnMessage(conn, nowUnix)

	} else if n == 0 {

		// 处理 RDHUP事件
		_ = conn.handleClose()
	} else {
		conn.handleError(conn.Fd())
	}

	// 读取数据

	// 返回0(EOF), 则关闭连接

	// 返回-1， 表示有错误发生

	return nil
}

// 2. 处理写
// ??? 何时激活读写
//
func (conn *Connection) handleWrite(fd int) error {
	fmt.Println("handleWrite: fd:", fd)

	// 1. 如果缓冲区中没有可读数据，则直接写入fd

	// 2. 否则说明写入过，则追加到缓冲区后边

	// redis 是使用链表把缓冲区拉起来
	// muduo 采用了上边说的策略
	// mdgo 如何处理呢???

	n, err := syscall.Write(conn.Fd(), conn.OutBuf.PeekAll())
	if err != nil {
		if err == syscall.EAGAIN { /// 之后，再次处理
			fmt.Println("happen EAGAIN")
			return nil
		}
		// 处理HUP事件
		return conn.handleClose()
	}

	if n == conn.OutBuf.ReadableBytes() {

		// 已经写完了
		// 则取消写事件
		// 激活读事件
		_ = conn.eventLoop.EnableRead(conn.Fd())

		// cb4
		// 这是缓冲区中，数据写完
		conn.cb.OnWriteComplete()
	}

	conn.OutBuf.Retrieve(n)

	return nil
}

// 3. 处理关闭
//
func (conn *Connection) handleClose() error {

	if conn.connected.Get() {
		conn.connected.Set(false)

		conn.eventLoop.DeleteInLoop(conn.Fd()) //

		// cb 3
		conn.cb.OnClose()

		/// 何时使用优雅关闭
		if err := syscall.Close(conn.Fd()); err != nil {
			fmt.Println("close fd err: ", err)
			return err
		}
	}
	return nil
}

// 4. 处理错误

func (conn *Connection) handleError(fd int) {
	nerr, err := syscall.GetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_ERROR)
	if err != nil {
		fmt.Println("TcpConnection::handleError => fd: ", fd, "; err: ", os.NewSyscallError("getsockopt", err))
		return
	}

	osErr := syscall.Errno(nerr)
	fmt.Println("TcpConnection::handleError => fd: ", fd, "; err: ", osErr.Error())

	// 这里真的有错误发生了，应该处理错误了
	// 直接退出程序???
	// 还是关闭连接???
	//
	//
	return
}

// 平滑关闭
func (conn *Connection) ShutdownWrite() error {
	conn.connected.Set(false)
	return syscall.Shutdown(conn.Fd(), syscall.SHUT_WR)
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
