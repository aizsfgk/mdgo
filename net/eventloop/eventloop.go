package eventloop

import (
	"fmt"

	"github.com/aizsfgk/mdgo/net/poller"
)

type Socket interface {
	Close() error
}

type EventLoop struct {
	Poll *poller.Poller
}


func New() (el *EventLoop, err error) {
	return
}

func (el *EventLoop) AddSocketAndEnableRead(fd int, s Socket) error {
	return nil
}

func (el *EventLoop) Stop() {
	return
}

func (el *EventLoop) Loop() {
	fmt.Println("eventLoop Loop...")
	return
}