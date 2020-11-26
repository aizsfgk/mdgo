package listener

import (
	"fmt"
	"github.com/aizsfgk/mdgo/net/event"
	"net"
	"os"
	"syscall"

	mdgoErr "github.com/aizsfgk/mdgo/net/error"
	"github.com/aizsfgk/mdgo/net/eventloop"
)

type HandlerConnFunc func(fd int, sa syscall.Sockaddr) error

type Listener struct {
	file *os.File
	listenFd int
	handleC HandlerConnFunc
	listener net.Listener
	loop *eventloop.EventLoop
}

func New(network, addr string, reusePort bool, loop *eventloop.EventLoop, handleConn HandlerConnFunc) (*Listener, error) {
	var (
		listener net.Listener
		err error
	)
	fmt.Println("network:", network)
	fmt.Println("addr:", addr)
	listener, err = net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	l, ok := listener.(*net.TCPListener)
	if !ok {
		return nil, mdgoErr.ListenerIsNotTcp
	}

	file, err := l.File()  /// 这里会发生FD dup; 文件描述符+1 TODO
	fmt.Printf("file:%#v\n", *file)
	if err != nil {
		return nil, err
	}

	fd := int(file.Fd())
	fmt.Println("New - fd: ", fd)
	if err = syscall.SetNonblock(fd, true); err != nil { // 设置了非阻塞
		return nil, err
	}

	return &Listener{
		file: file,
		listenFd: fd,
		handleC: handleConn,
		listener: listener,
		loop: loop,
	}, nil
}

func (l *Listener) HandleEvent(fd int, eve event.Event) error {
	if eve & event.EventRead != 0 {
		connFd, sa, err := syscall.Accept4(fd, syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC)
		if err != nil {
			if err != syscall.EAGAIN {
				fmt.Println("accept4-err:", err.Error())
				return nil
			}
			return err
		}
		fmt.Println("connFd: ", connFd)
		return l.handleC(connFd, sa)
	}
	return nil
}

func (l *Listener) Fd() int {
	return l.listenFd
}

func (l *Listener) Close() error {
	l.loop.DeleteInLoop(l.listenFd)
	return nil
}

