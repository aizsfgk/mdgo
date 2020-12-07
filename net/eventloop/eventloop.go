package eventloop

import (
	"fmt"

	"github.com/aizsfgk/mdgo/net/event"
	"github.com/aizsfgk/mdgo/base/atomic"
	"github.com/aizsfgk/mdgo/net/poller"
)

type SocketCtx interface {
	Close() error
	HandleEvent(eve event.Event, nowUnix int64) error
}

type EventLoop struct {
	Poll *poller.Poller
	socketCtx map[int]SocketCtx // 连接和对应的处理程序

	looping atomic.Bool
	quit atomic.Bool
	eventHandling atomic.Bool
	LoopId int
}


func New() (el *EventLoop, err error) {
	fmt.Println("create Poller")
	p, err := poller.Create()
	if err != nil {
		return nil, err
	}
	return &EventLoop{
		Poll:    p,
		socketCtx: make(map[int]SocketCtx, 1024),
	}, nil
}

func (el *EventLoop) AddSocketAndEnableRead(fd int, sc SocketCtx) error {


	var err error
	el.socketCtx[fd] = sc

	if err = el.Poll.Add(fd, event.EventRead); err != nil {
		fmt.Println("el.Poll.Add err: ", err)
		el.Poll.Del(fd)
		return err
	}
	return nil
}

func (el *EventLoop) Stop() error {
	for fd, sc := range el.socketCtx {
		if err := sc.Close(); err != nil {
			fmt.Println("Stop-err: ", err)
		}
		delete(el.socketCtx, fd)
	}
	return el.Poll.Close()
}

func (el *EventLoop) debugPrintf(evs *[]event.Ev) {
	fmt.Printf("\n==========================\n")
	fmt.Printf(" revent-print: \n")
	for _, ev := range *evs {
		if ev.Fd > 0 {
			fmt.Printf("fd: %d => events: %s\n", ev.Fd, ev.RString())
		}
	}
	fmt.Printf("==========================\n")
}

// 开启事件循环
func (el *EventLoop) Loop() {
	fmt.Println("eventLoop Loop begin; idx： ", el.LoopId)

	el.looping.Set(true)

	for !el.quit.Get() {
		activeConn := make([]event.Ev, poller.WaitEventsBegin)
		nowUnix, n := el.Poll.Poll(1000, &activeConn)
		fmt.Println("idx: ", i, " 返回")

		if n > 0 {
			el.debugPrintf(&activeConn)

			el.eventHandling.Set(true) // 开启事件处理
			for _, curEv := range activeConn {
				if sc, ok := el.socketCtx[curEv.Fd]; ok {
					err := sc.HandleEvent(curEv.Revent, nowUnix)
					if err != nil {
						fmt.Println("handler activeConn: err: ", err)
					}
				}
			}
			el.eventHandling.Set(false)
		}
	}

	fmt.Println("eventLoop Loop end...")
	el.looping.Set(false)
	return
}

func (el *EventLoop) EnableRead(fd int) error {
	return el.Poll.EnableRead(fd)
}

func (el *EventLoop) EnableWrite(fd int) error {
	return el.Poll.EnableWrite(fd)
}

func (el *EventLoop) EnableReadWrite(fd int) error {
	return el.Poll.EnableReadWrite(fd)
}

func (el *EventLoop) DeleteInLoop(fd int) {
	// 剔除事件循环
	if err := el.Poll.Del(fd); err != nil {
		fmt.Println("[DeleteFdInLoop]", err)
	}
	delete(el.socketCtx, fd)
}