package connection

import (
	"bufio"
	"github.com/aizsfgk/mdgo/net/event"
	"net"
	"syscall"

	"github.com/aizsfgk/mdgo/net/eventloop"
)

type ConnState int

const (
	KConnecting    ConnState = 1
	KConnected     ConnState = 2
	KDisConnecting ConnState = 3
	KDisConnected  ConnState = 4
)

type Connection struct {
	connFd int
	state ConnState
	br bufio.Reader
	bw bufio.Writer

	peerAddr net.Addr
	eventLoop *eventloop.EventLoop
}


func New(fd int, loop *eventloop.EventLoop, sa syscall.Sockaddr) (*Connection, error) {
	return nil, nil
}


func (conn *Connection) Fd() int {
	return conn.connFd
}

func (conn *Connection) Close() error {
	return nil
}

func (conn *Connection) HandleEvent(fd int, eve event.Event) error {
	return nil
}