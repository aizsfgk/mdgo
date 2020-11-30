package ringbuffer

const (
	cheapPrepend = 8
	kInitSize = 1024
)

type FixBuffer struct {
	preSize int
	ri int
	wi int
	buf []byte
}

func NewFixBuffer() *FixBuffer {
	return &FixBuffer{
		ri:  cheapPrepend,
		wi:  cheapPrepend,
		buf: make([]byte, kInitSize),
	}
}

func (f *FixBuffer) swap(o *FixBuffer) {
	f.buf, o.buf = o.buf, f.buf
	f.ri, o.ri = o.ri, f.ri
	f.wi, o.wi = o.wi, f.wi
}

func (f *FixBuffer) readableBytes() int {
	return f.wi - f.ri
}

func (f *FixBuffer) writableBytes() int {
	return len(f.buf) - f.wi
}

func (f *FixBuffer) prependableBytes() int {
	return f.ri
}


// *************** retrieve *************** //
func (f *FixBuffer) retrieve(len int) {
	if len < f.readableBytes() {
		f.ri += len
	} else {
		f.retrieveAll()
	}
}

func (f *FixBuffer) retrieveAll() {
	f.ri, f.wi = cheapPrepend, cheapPrepend
}

func (f *FixBuffer) retrieveInt64() {
	f.retrieve(8)
}

func (f *FixBuffer) retrieveInt32() {
	f.retrieve(4)
}

func (f *FixBuffer) retrieveInt16() {
	f.retrieve(2)
}

func (f *FixBuffer) retrieveInt8() {
	f.retrieve(1)
}



// ************** write / append **************** //
func (f *FixBuffer) append(b []byte) {
	f.buf = append(f.buf[:len(f.buf)], b...)
}

func (f *FixBuffer) unwrite(len int) {
	if len > f.readableBytes() {
		return
	}
	f.wi -= len
}

func (f *FixBuffer) appendInt64() {

}

func (f *FixBuffer) appendInt32() {

}

func (f *FixBuffer) appendInt16() {

}

func (f *FixBuffer) appendInt8() {

}

// ******************* read ******************* //


// ******************* peek ******************** //


// ******************* prepend **************** //











