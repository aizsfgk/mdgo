package event

import "bytes"

type Event uint32

const (
	// 这些定义的事件 - 是给eventloop用的
	EventRead  Event = 0x01
	EventWrite Event = 0x02
	EventErr   Event = 0x80
	EventNone        = 0
)

type Ev struct {
	Fd     int
	Event  Event
	Revent Event
}

func (e *Ev) RString() (out string) {
	var buf bytes.Buffer
	if e.Revent & EventRead != 0 {
		buf.WriteString("READ, ")
	}

	if e.Revent & EventWrite != 0 {
		buf.WriteString("WRITE, ")
	}

	if e.Revent & EventErr != 0 {
		buf.WriteString("ERROR, ")
	}
	out = buf.String()
	return
}