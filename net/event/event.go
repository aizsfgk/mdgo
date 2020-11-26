package event

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
