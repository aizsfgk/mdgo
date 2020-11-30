package poller

import (
	"fmt"
	"syscall"

	"github.com/aizsfgk/mdgo/base/atomic"
	"github.com/aizsfgk/mdgo/net/event"
	mdgoErr "github.com/aizsfgk/mdgo/net/error"
)

const (
	/*
	   1. 水平触发: 只要接收缓冲区中有数据，就会通知一次
	   2. 边缘触发: 接收缓冲区从空到非空，才会通知一次
			-syscall.EPOLLET
	 */
	readEvent = syscall.EPOLLIN
	writeEvent = syscall.EPOLLOUT
	errorEvent = syscall.EPOLLERR // epoll 默认会注册这种事件
)


type Poller struct {
	epFd int            // epFd
	eventFd int         // wakeup用, FIXME
	running atomic.Bool
}

// 创建
func Create() (*Poller, error) {
	epFd, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		fmt.Println("Create-err: ", err)
		syscall.Close(epFd)
		return nil, err
	}
	return &Poller{
		epFd:    epFd,
		eventFd: 0,
	},nil
}

func (p *Poller) Close() error {
	_ = syscall.Close(p.epFd)
	return nil
}

// ***************** 操作 - 私有方法 ***************** //
// 增加事件
func (p *Poller) add(fd int, events uint32) error {
	return syscall.EpollCtl(p.epFd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{
		Events: events,
		Fd:     int32(fd),
	})
}

// 修改事件
func (p *Poller) mod(fd int, events uint32) error {
	return syscall.EpollCtl(p.epFd, syscall.EPOLL_CTL_MOD, fd, &syscall.EpollEvent{
		Events: events,
		Fd:     int32(fd),
	})
}

// 删除事件
func (p *Poller) Del(fd int) error {
	return syscall.EpollCtl(p.epFd, syscall.EPOLL_CTL_DEL, fd, nil)
}

func (p *Poller) Add(fd int, eve event.Event) error {
	var events uint32

	if eve & event.EventRead != 0 {
		events |= readEvent
		fmt.Println("add readEvent")
		fmt.Println("events:", events)
		return p.add(fd, events)
	}

	if eve & event.EventWrite != 0 {
		events |= writeEvent
		return p.add(fd, events)
	}

	return mdgoErr.EventIsNil
}


// **************** 激活 - 共有 ***************** //
func (p *Poller) EnableRead(fd int) error {
	return p.mod(fd, readEvent)
}

func (p *Poller) EnableWrite(fd int) error {
	return p.mod(fd, writeEvent)
}

func (p *Poller) EnableReadWrite(fd int) error {
	return p.mod(fd, readEvent|writeEvent)
}

func (p *Poller) Poll(msec int, acp *[]*event.Ev) {
	events := make([]syscall.EpollEvent, WaitEventsBegin)

	n, err := syscall.EpollWait(p.epFd, events, msec) // 事件就绪
	if err != nil && err != syscall.EAGAIN {
		fmt.Println("epollWait-err: ", err)
		return
	}

	var evp event.Ev
	for i:=0; i<n; i++ {
		var rEvent event.Event
		if events[i].Events & errorEvent != 0 {
			rEvent |= event.EventErr
		}
		if events[i].Events & readEvent != 0 {
			rEvent |= event.EventRead
		}
		if events[i].Events & writeEvent != 0 {
			rEvent |= event.EventWrite
		}

		evp.Revent = rEvent
		evp.Fd = int(events[i].Fd)

		*acp = append(*acp, &evp)
	}

	if len(events) == n {
		WaitEventsBegin = 2 * WaitEventsBegin
	}
	return
}




