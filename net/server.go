package net

import (
	"fmt"
	"runtime"
	"syscall"

	"github.com/aizsfgk/mdgo/base/atomic"
	"github.com/aizsfgk/mdgo/net/connection"
	mdgoErr "github.com/aizsfgk/mdgo/net/error"
	"github.com/aizsfgk/mdgo/net/eventloop"
	"github.com/aizsfgk/mdgo/net/listener"
)

type Handler interface {
	OnEventLoopInit(conn *connection.Connection)
	OnConnection(conn *connection.Connection)
	connection.Callback
}

type Codec interface {
	Pack(b []byte)
	Unpack(b []byte) interface{}
}


// 编解码和服务器走
type Server struct {
	started atomic.Bool
	option  *Option
	handler Handler
	codec Codec

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

	if serv.option.NumLoop > runtime.NumCPU() {
		serv.option.NumLoop = runtime.NumCPU()
	}

	fmt.Println("serv.option.NumLoop: ", serv.option.NumLoop)
	if serv.option.NumLoop > 0 {
		wloops := make([]*eventloop.EventLoop, serv.option.NumLoop)
		for i := 0; i < serv.option.NumLoop; i++ {
			loop, err := eventloop.New()
			loop.LoopId = i
			if err != nil {
				fmt.Println("wloops-err:", wloops)
				for j := 0; j < i; j++ {
					wloops[i].Stop()
				}
				return nil, err
			}
			wloops[i] = loop
		}
		serv.worksLoop = wloops
	}

	return
}

func (serv *Server) nextLoop() *eventloop.EventLoop {
	if serv.worksNum == 0 {
		return serv.mainLoop
	}
	loop := serv.worksLoop[serv.nextLoopIndex]
	serv.nextLoopIndex = (serv.nextLoopIndex + 1) % len(serv.worksLoop)
	return loop
}

func (serv *Server) handleNewConnection(fd int, sa syscall.Sockaddr) error {
	fmt.Println("handleNewConnection start")
	loop := serv.nextLoop()

	// 新建连接
	//
	// Connection 是对 TCP连接的抽象
	conn, err := connection.New(fd, loop, sa, serv.handler)

	if err != nil {
		fmt.Println("handleNewConnection:err: ", err)
		return err
	}

	// cb1 ： 执行回调
	serv.handler.OnConnection(conn)


	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_SNDBUF, 10)
	if err != nil {
		fmt.Println("set sndBuf err1: ", err)
	}

	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_RCVBUF, 10)
	if err != nil {
		fmt.Println("set sndBuf err2: ", err)
	}

	return loop.AddSocketAndEnableRead(fd, conn)
}

func (serv *Server) Start() (err error) {
	fmt.Println("server start")

	serv.mainLoop.Loop()

	return
}

func (serv *Server) Stop() {
	return
}
