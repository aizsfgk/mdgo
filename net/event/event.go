package event

import "bytes"

type Event uint32

const (
	EventNone  Event = 0
	EventRead  Event = 0x01
	EventWrite Event = 0x02
	EventError Event = 0x04
)

type EventHolder struct {
	Fd     int   // fd
	Event  Event // 关注的事件
	Revent Event // 就绪的事件
}

func (e *EventHolder) Event2String() (out string) {
	var buf bytes.Buffer

	if e.Revent&EventRead != 0 {
		buf.WriteString("READ, ")
	}
	if e.Revent&EventWrite != 0 {
		buf.WriteString("WRITE, ")
	}
	if e.Revent&EventError != 0 {
		buf.WriteString("ERROR, ")
	}

	out = buf.String()
	return
}
