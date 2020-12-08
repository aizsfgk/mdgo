// +build linux

package poller

import (
	"fmt"
	"syscall"
	"time"

	mdgoErr "github.com/aizsfgk/mdgo/net/errors"
	"github.com/aizsfgk/mdgo/net/event"
	"github.com/aizsfgk/mdgo/base/atomic"
)

const (
	/*
	   1. 水平触发: 只要接收缓冲区中有数据，就会通知一次
	   2. 边缘触发: 接收缓冲区从空到非空，才会通知一次
			-syscall.EPOLLET

	EPOLLIN : 可读
	EPOLLOUT： 可写
	EPOLLRDHUP : 对端关闭或者半关闭（写） since linux 2.6.17
	EPOLLPRI : 带外数据可读

	EPOLLERR： fd相关的错误事件事件发生（总是监听）
	EPOLLHUP:  相关的FD上发生Hang Up事件[ shutdown(wr) shutdown(rd) close(conn) ]

	// 当发生这些错误的时候，需要从epoll中删除这些Fd,并关闭连接

	EPOLLET: 边缘触发
	EPOLLONESHOT： 只通知一次 since 2.6.2

	https://elixir.bootlin.com/linux/v4.19/source/net/ipv4/tcp.c#L524

	if (sk->sk_shutdown == SHUTDOWN_MASK || state == TCP_CLOSE) # 意味着如果你关闭了连接，还监听该FD，则会返回EPOLLHUP
			mask |= EPOLLHUP;
	if (sk->sk_shutdown & RCV_SHUTDOWN)
		mask |= EPOLLIN | EPOLLRDNORM | EPOLLRDHUP;

	if (sk->sk_err || !skb_queue_empty(&sk->sk_error_queue))    # 错误发生的条件
			mask |= EPOLLERR;

	 */
	readEvent  = syscall.EPOLLIN | syscall.EPOLLPRI | syscall.EPOLLRDHUP
	writeEvent = syscall.EPOLLOUT | syscall.EPOLLHUP
	errorEvent = syscall.EPOLLERR // 发生了真实的错误
)


type Poller struct {
	running atomic.Bool
	epFd int
	events []syscall.EpollEvent
}

// 创建
func Create() (*Poller, error) {
	epFd, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC) // 为何要使用这些标志
	if err != nil {
		fmt.Println("Create-err: ", err)
		syscall.Close(epFd)
		return nil, err
	}
	return &Poller{
		epFd:    epFd,
		events:  make([]syscall.EpollEvent, WaitEventsBegin),
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

/*
	就绪事件如何暴露出来???这是一个值得思考的问题

 */
func (p *Poller) Poll(msec int, acp *[]event.EventHolder) (int64, int) {

	n, err := syscall.EpollWait(p.epFd, p.events, msec) // 事件就绪
	nowUnix := time.Now().Unix()
	fmt.Println("就绪事件个数：", n)
	if err != nil{
		// EAGAIN : 非阻塞模式，EAGIN
		// EINTR: 系统繁忙被打断
		if err == syscall.EAGAIN || err == syscall.EINTR {
			return nowUnix, 0
		}
		fmt.Println("epollWait-err: ", err)
		return nowUnix, 0
	}

	var evHolder event.EventHolder
	for i:=0; i<n; i++ {
		retEvent := event.EventNone
		if p.events[i].Events & errorEvent != 0 {
			retEvent |= event.EventError
		}
		if p.events[i].Events & readEvent != 0 {
			retEvent |= event.EventRead
		}
		if p.events[i].Events & writeEvent != 0 {
			retEvent |= event.EventWrite
		}
		evHolder.Revent = retEvent
		evHolder.Fd = int(p.events[i].Fd)

		(*acp)[i] = evHolder
	}

	if len(p.events) == n {
		p.events = make([]syscall.EpollEvent, 2 * WaitEventsBegin)
	}

	return nowUnix, n
}




