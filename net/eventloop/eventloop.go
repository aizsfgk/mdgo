package eventloop

import (
	"fmt"

	"github.com/aizsfgk/mdgo/net/event"
	"github.com/aizsfgk/mdgo/base/atomic"
	"github.com/aizsfgk/mdgo/net/poller"
)

type SocketCtx interface {
	Close() error
	HandleEvent(fd int, eve event.Event) error
}

type EventLoop struct {
	Poll *poller.Poller
	socketCtx map[int]SocketCtx

	looping atomic.Bool
	quit atomic.Bool
	eventHandling atomic.Bool
}


func New() (el *EventLoop, err error) {
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
	fmt.Println("AddSocketAndEnableRead-fd: ", fd)
	var err error
	el.socketCtx[fd] = sc

	if err = el.Poll.Add(fd, event.EventRead); err != nil {
		fmt.Println("el.Poll.Add err: ", err)
		el.Poll.Del(fd)
		return err
	}
	fmt.Println("add yes")
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

// 开启事件循环
func (el *EventLoop) Loop() {
	fmt.Println("eventLoop Loop begin ...")

	el.looping.Set(true)

	activeConn := make([]*event.Ev, 0, poller.WaitEventsBegin)

	for !el.quit.Get() {
		el.Poll.Poll(10000, &activeConn)

		if len(activeConn) > 0 {
			fmt.Println("activeConn-len>0")
			el.eventHandling.Set(true)

			for _, curEv := range activeConn {
				if sc, ok := el.socketCtx[curEv.Fd]; ok {
					err := sc.HandleEvent(curEv.Fd, curEv.Revent)
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
	if err := el.Poll.Del(fd); err != nil {
		fmt.Println("[DeleteFdInLoop]", err)
	}
	delete(el.socketCtx, fd)
}