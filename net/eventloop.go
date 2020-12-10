package net

import (
	"fmt"

	"github.com/aizsfgk/mdgo/base/atomic"
	_const "github.com/aizsfgk/mdgo/net/const"
	"github.com/aizsfgk/mdgo/net/event"
	"github.com/aizsfgk/mdgo/net/poller"
)

// socket context
// 1. listener
// 2. connection
type SocketContext interface {
	Close() error
	HandleEvent(eve event.Event, nowUnix int64) error
}

// 事件循环
type EventLoop struct {
	Poll          *poller.Poller        // net poller
	LoopId        string                // identify
	socketCtx     map[int]SocketContext // fd <-> SocketContext
	quit          atomic.Bool           // is quit
	eventHandling atomic.Bool           // is handle event
}

// New/Loop/Stop
// Enable Read/Write/ReadWrite
// DeleteInLoop

// new EventLoop
func NewEventLoop() (el *EventLoop, err error) {
	poll, err := poller.Create()
	if err != nil {
		return nil, err
	}
	return &EventLoop{
		Poll:      poll,
		socketCtx: make(map[int]SocketContext, _const.SocketContextSize),
	}, nil
}

// first add and enable read
func (el *EventLoop) AddSocketAndEnableRead(fd int, sckCtx SocketContext) error {
	var err error

	el.socketCtx[fd] = sckCtx
	if err = el.Poll.Add(fd, event.EventRead); err != nil {
		_ = el.Poll.Del(fd)
		return err
	}
	return nil
}

// stop eventLoop
func (el *EventLoop) Stop() error {
	for fd, sc := range el.socketCtx {
		if err := sc.Close(); err != nil {
			fmt.Println("sc close err")
		}
		delete(el.socketCtx, fd)
	}
	return el.Poll.Close()
}

// debugPrintf
func (el *EventLoop) debugPrintf(evhs *[]event.EventHolder) {
	fmt.Println("==========================")
	fmt.Println("return Event: ")
	for _, evh := range *evhs {
		if evh.Fd > 0 {
			fmt.Printf("fd: %d => events: %s\n", evh.Fd, evh.Event2String())
		}
	}
	fmt.Println("==========================")
}

// 开启事件循环
func (el *EventLoop) Loop() {
	fmt.Println("<<< eventLoop Loop begin; LoopId: ", el.LoopId, ">>>")

	for !el.quit.Get() {
		activeEvents := make([]event.EventHolder, poller.WaitEventsBegin)
		nowUnix, n := el.Poll.Poll(_const.PollWaitMillisecond, &activeEvents)

		fmt.Println("return eventLoop; LoopId: ", el.LoopId, "; unix timestamp: ", nowUnix)

		if n > 0 {
			el.debugPrintf(&activeEvents)

			el.eventHandling.Set(true)
			for _, curEvent := range activeEvents {
				if sc, ok := el.socketCtx[curEvent.Fd]; ok {
					err := sc.HandleEvent(curEvent.Revent, nowUnix)
					if err != nil {
						fmt.Println("sc.HandleEvent err: ", err.Error())
					}
				}
			}
			el.eventHandling.Set(false)
		}
	}

	fmt.Println("<<< eventLoop loop end >>>")
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
	// delete from eventLoop Poll
	if err := el.Poll.Del(fd); err != nil {
		fmt.Println("[DeleteFdInLoop]", err)
	}

	// delete from socketContext
	delete(el.socketCtx, fd)
}
