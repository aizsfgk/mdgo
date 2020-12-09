package net

import (
	"fmt"
	"net"
	"os"
	"syscall"

	mdgoErr "github.com/aizsfgk/mdgo/net/errors"
	"github.com/aizsfgk/mdgo/net/event"
)

type HandlerConnFunc func(fd int, sa syscall.Sockaddr) error

type Listener struct {
	listenFd      int             // 监听套接字
	file          *os.File        // dupFd
	handleNewConn HandlerConnFunc // new acceptFd coming to handle
	listener      net.Listener    // net.Listener
	loop          *EventLoop      // pointer main eventloop
}

func NewListener(network, addr string, reusePort bool, loop *EventLoop, handleConn HandlerConnFunc) (*Listener, error) {
	var (
		listener    net.Listener
		tcpListener *net.TCPListener
		err         error
		ok          bool
	)

	listener, err = net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	tcpListener, ok = listener.(*net.TCPListener)
	if !ok {
		return nil, mdgoErr.ListenerIsNotTcp
	}

	file, err := tcpListener.File() // File() 会发生 dupFd(), 由于是listenerFd,所以一个服务只会多一个
	if err != nil {
		return nil, err
	}

	fd := int(file.Fd())
	fmt.Println("*** new listenerFd: ", fd, "***")
	if err = syscall.SetNonblock(fd, true); err != nil {
		return nil, err
	}

	return &Listener{
		file:          file,
		listenFd:      fd,
		handleNewConn: handleConn,
		listener:      listener,
		loop:          loop,
	}, nil
}

func (l *Listener) HandleEvent(eve event.Event, nowUnix int64) error {
	// listenerFd only handle EventRead event
	if eve&event.EventRead != 0 {
		connFd, sa, err := syscall.Accept4(l.listenFd, syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC)
		if err != nil {
			if err == syscall.EAGAIN { // due to listenFd is nonblock, so accept can return EAGAIN
				return nil
			}
			return err
		}
		fmt.Println("*** new connFd: ", connFd, "***")
		// start handle new connection
		return l.handleNewConn(connFd, sa)
	}
	return nil
}

func (l *Listener) Fd() int {
	return l.listenFd
}

func (l *Listener) Close() error {
	l.loop.DeleteInLoop(l.listenFd)
	return syscall.Close(l.listenFd)
}
