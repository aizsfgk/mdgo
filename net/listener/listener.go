package listener

import (
	"fmt"

	mdgoErr "github.com/aizsfgk/mdgo/net/error"
	"net"
	"os"
	"syscall"

	"github.com/aizsfgk/mdgo/net/eventloop"
)

type HandlerConnFunc func(fd int, unix syscall.Sockaddr)

type Listener struct {
	file *os.File
	fd int
	handleC HandlerConnFunc
	listener net.Listener
	loop *eventloop.EventLoop
}

func New(network, addr string, reusePort bool, loop *eventloop.EventLoop, handleConn HandlerConnFunc) (*Listener, error) {
	var (
		listener net.Listener
		err error
	)

	listener, err = net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	l, ok := listener.(*net.TCPListener)
	if !ok {
		return nil, mdgoErr.ListenerIsNotTcp
	}

	file, err := l.File()  /// 这里会发生FD dup; 文件描述符+1 TODO
	fmt.Printf("file:%#v\n", file)
	if err != nil {
		return nil, err
	}

	fd := int(file.Fd())
	fmt.Println("fd: ", fd)
	if err = syscall.SetNonblock(fd, true); err != nil {
		return nil, err
	}

	return &Listener{
		file: file,
		fd: fd,
		handleC: handleConn,
		listener: listener,
		loop: loop,
	}, nil
}

func (l *Listener) Fd() int {
	return l.fd
}

func (l *Listener) Close() error {
	return nil
}

