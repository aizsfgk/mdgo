package net

import (
	"fmt"
	"runtime"
	"syscall"
	mdgoErr "github.com/aizsfgk/mdgo/net/error"

	"github.com/aizsfgk/mdgo/net/listener"
	"github.com/aizsfgk/mdgo/base/atomic"
	"github.com/aizsfgk/mdgo/net/connection"
	"github.com/aizsfgk/mdgo/net/eventloop"
)

type Handler interface {
	OnEventLoopInit(conn *connection.Connection)
	OnConnection(conn *connection.Connection)
	OnMessage(conn *connection.Connection)
	OnWriteComplete(conn *connection.Connection)
	OnClose(conn *connection.Connection)
}

type Server struct {
	started atomic.Bool
	option  *Option
	handler Handler

	mainLoop      *eventloop.EventLoop
	worksNum      int
	worksLoop     []*eventloop.EventLoop
	nextLoopIndex int
}

func NewServer(handler Handler, optionCbs ...OptionCallback) (serv *Server, err error) {
	if handler == nil {
		err = mdgoErr.HandlerIsNil
		return
	}
	option := newOption(optionCbs...)
	serv = new(Server)

	serv.handler = handler
	serv.option = option
	serv.mainLoop, err = eventloop.New()
	if err != nil {
		_ = serv.mainLoop.Stop
		return nil, err
	}

	l, err := listener.New(serv.option.Network, serv.option.Addr, serv.option.ReusePort, serv.mainLoop, serv.handleNewConnection)
	if err != nil {
		return nil, err
	}
	if err = serv.mainLoop.AddSocketAndEnableRead(l.Fd(), l); err != nil {
		return nil, err
	}

	if serv.option.NumLoop <= 0 {
		serv.option.NumLoop = runtime.NumCPU()
	}

	wloops := make([]*eventloop.EventLoop, serv.option.NumLoop)
	for i:=0; i<serv.option.NumLoop; i++ {
		loop, err := eventloop.New()
		if err != nil {
			for j:=0; j<i; j++ {
				wloops[i].Stop()
			}
			return nil, err
		}
		wloops[i] = loop
	}
	serv.worksLoop = wloops

	return
}

func (serv *Server) handleNewConnection(fd int, sa syscall.Sockaddr) {

}

func (serv *Server) Start() (err error) {
	fmt.Println("server start")
	return
}

func (serv *Server) Stop() {
	return
}
