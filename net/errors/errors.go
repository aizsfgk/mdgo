package errors

import (
	"errors"
)

var (
	HandlerIsNil = errors.New("server handler is nil")
	ListenerIsNotTcp = errors.New("listener is not tcp")
	EventIsNil = errors.New("event is nil")
	ErrConnectionClosed = errors.New("connection closed")
)
