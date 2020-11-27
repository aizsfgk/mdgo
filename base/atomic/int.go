package atomic

import "sync/atomic"

// 32
type Int32 struct {
	v int32
}

func (a *Int32) Add(i int32) int32 {
	return atomic.AddInt32(&a.v, i)
}

func (a *Int32) Swap(i int32) int32 {
	return atomic.SwapInt32(&a.v, i)
}

func (a *Int32) Get() int32 {
	return atomic.LoadInt32(&a.v)
}

// 64
type Int64 struct {
	v int64
}

func (a *Int64) Add(i int64) int64 {
	return atomic.AddInt64(&a.v, i)
}

func (a *Int64) Swap(i int64) int64 {
	return atomic.SwapInt64(&a.v, i)
}

func (a *Int64) Get() int64 {
	return atomic.LoadInt64(&a.v)
}
