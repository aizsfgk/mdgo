package atomic

import "sync/atomic"

type Bool struct {
	b int32
}

func (a *Bool) Set(b bool) bool {
	var vNew int32
	if b {
		vNew = 1
	}
	return atomic.SwapInt32(&a.b, vNew) == 1
}

func (a *Bool) Get() bool {
	return atomic.LoadInt32(&a.b) == 1
}
