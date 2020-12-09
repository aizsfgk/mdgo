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
	listenFd int
	file *os.File
	handleC HandlerConnFunc
	listener net.Listener
	loop *EventLoop
}

func NewListener(network, addr string, reusePort bool, loop *EventLoop, handleConn HandlerConnFunc) (*Listener, error) {
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
	if err != nil {
		return nil, err
	}

	fd := int(file.Fd())
	fmt.Println("New - fd: ", fd)
	if err = syscall.SetNonblock(fd, true); err != nil { // 设置了非阻塞
		return nil, err
	}

	/*
	没有生效
	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_SNDBUF, 10)
	if err != nil {
		fmt.Println("set sndBuf err1: ", err)
	}

	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_RCVBUF, 10)
	if err != nil {
		fmt.Println("set sndBuf err2: ", err)
	}
	 */

	return &Listener{
		file: file,
		listenFd: fd,
		handleC: handleConn,
		listener: listener,
		loop: loop,
	}, nil
}

func (l *Listener) HandleEvent(eve event.Event, nowUnix int64) error {
	fmt.Println("listener HandleEvent")
	// 监听的套接字，只关注可读事件
	if eve & event.EventRead != 0 {
		connFd, sa, err := syscall.Accept4(l.listenFd, syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC)
		if err != nil {
			if err == syscall.EAGAIN { // EAGAIN : 表示资源临时不可用
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

