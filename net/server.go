package net

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"syscall"

	"github.com/aizsfgk/mdgo/base/atomic"
	mdgoErr "github.com/aizsfgk/mdgo/net/errors"
)

// 处理句柄
type Handler interface {
	OnEventLoopInit(conn *Connection)
	OnConnection(conn *Connection)
	Callback
}

// 编解码器
type Codec interface {
	Pack(b []byte)
	Unpack(b []byte) interface{}
}

// 服务器
type Server struct {
	started       atomic.Bool    // 标明服务器是否启动
	option        *Option        // 配置选项
	handler       Handler        // 回调句柄
	codec         Codec          // 编解码器 ??? 是否可以放到 EventLoop 减少锁开销
	mainLoop      *EventLoop     // mainReactor
	workLoops     []*EventLoop   // subReactor
	nextLoopIndex int            // workLoop索引
	wg            sync.WaitGroup // 同步
}

// 新建服务器
func NewServer(handler Handler, optionCbs ...OptionCallback) (serv *Server, err error) {
	if handler == nil {
		err = mdgoErr.HandlerIsNil
		return
	}

	// new server
	serv = new(Server)
	serv.handler = handler
	serv.option = newOption(optionCbs...)
	serv.mainLoop, err = NewEventLoop()
	if err != nil {
		_ = serv.mainLoop.Stop
		return nil, err
	}
	serv.mainLoop.LoopId = "mainReactor"

	// new listener
	listener, err := NewListener(serv.option.Network, serv.option.Addr, serv.option.ReusePort, serv.mainLoop, serv.handleNewConnection)
	if err != nil {
		return nil, err
	}
	if err = serv.mainLoop.AddSocketAndEnableRead(listener.Fd(), listener); err != nil {
		return nil, err
	}

	if serv.option.NumLoop > runtime.NumCPU() {
		serv.option.NumLoop = runtime.NumCPU()
	}

	// new sub eventLoop
	if serv.option.NumLoop > 0 {
		subLoops := make([]*EventLoop, serv.option.NumLoop)
		for i := 0; i < serv.option.NumLoop; i++ {
			loop, err := NewEventLoop()
			loop.LoopId = "subReactor-IDX-" + strconv.Itoa(i)
			if err != nil {
				fmt.Println("NewEventLoop sub err: ", err.Error())
				for j := 0; j < i; j++ {
					_ = subLoops[i].Stop()
				}
				return nil, err
			}
			subLoops[i] = loop
		}
		serv.workLoops = subLoops
	}

	return
}

// 启动服务器
func (serv *Server) Start() (err error) {
	fmt.Println("Server Start ...")

	// subReactor Loop
	for i := 0; i < serv.option.NumLoop; i++ {
		serv.wg.Add(1)
		go func(idx int) {
			defer serv.wg.Done()
			serv.workLoops[idx].Loop()
		}(i)
	}

	// mainReactor Loop
	serv.wg.Add(1)
	go func() {
		defer serv.wg.Done()
		serv.mainLoop.Loop()
	}()
	serv.wg.Wait()

	fmt.Println("Server Start End ...")
	return
}

// 停止服务器
func (serv *Server) Stop() {
	if err := serv.mainLoop.Stop(); err != nil {
		fmt.Println("mainLoop Stop err: ", err)
		return
	}

	for _, wlp := range serv.workLoops {
		_ = wlp.Stop()
	}
	return
}

// ******************** private method ******************** //

// 获取NextLoop
func (serv *Server) nextEventLoop() *EventLoop {
	if serv.option.NumLoop == 0 {
		return serv.mainLoop
	}
	loop := serv.workLoops[serv.nextLoopIndex]
	serv.nextLoopIndex = (serv.nextLoopIndex + 1) % len(serv.workLoops)
	return loop
}

// 新到连接处理
func (serv *Server) handleNewConnection(fd int, sa syscall.Sockaddr) error {

	// get next eventLoop
	loop := serv.nextEventLoop()

	// new connection
	conn, err := NewConnection(fd, loop, sa, serv.handler)
	if err != nil {
		fmt.Println("NewConnection err: ", err.Error())
		return err
	}

	// cb: OnConnection
	serv.handler.OnConnection(conn)

	// register event[Read]
	return loop.AddSocketAndEnableRead(fd, conn)
}
