package connection

import (
	"bufio"
	"net"

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
	sockFd int
	state ConnState
	br bufio.Reader
	bw bufio.Writer

	localAddr *net.Addr
	peerAddr net.Addr
	eventLoop *eventloop.EventLoop
}


func (conn *Connection) Fd() int {
	return conn.sockFd
}