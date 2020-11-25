package net

import "time"

type Option struct {
	Network string
	Addr    string

	NumLoop   int
	ReusePort bool
	KeepAlive time.Duration
}

type OptionCallback func(*Option)

func newOption(optCb ...OptionCallback) *Option {
	opt := Option{}

	for _, cb := range optCb {
		cb(&opt)
	}

	if len(opt.Network) == 0 {
		opt.Network = "tcp"
	}

	if len(opt.Addr) == 0 {
		opt.Addr = ":9292"
	}

	return &opt
}

func ReusePort(reusePort bool) OptionCallback {
	return func(o *Option) {
		o.ReusePort = reusePort
	}
}

func Network(n string) OptionCallback {
	return func(o *Option) {
		o.Network = n
	}
}

func Addr(addr string) OptionCallback {
	return func(o *Option) {
		o.Addr = addr
	}
}

func NumLoop(nl int) OptionCallback {
	return func(o *Option) {
		o.NumLoop = nl
	}
}

func KeepAlive(ka time.Duration) OptionCallback {
	return func(o *Option) {
		o.KeepAlive = ka
	}
}
