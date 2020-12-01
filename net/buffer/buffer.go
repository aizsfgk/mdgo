package buffer

import (
	"encoding/binary"
	"syscall"
	"unsafe"
)

const (
	cheapPrepend = 8
	kInitSize = 1024
)

// A buffer class modeled after org.jboss.netty.buffer.ChannelBuffer
//
// +-------------------+------------------+------------------+
// | prependable bytes |  readable bytes  |  writable bytes  |
// |                   |     (CONTENT)    |                  |
// +-------------------+------------------+------------------+
// |                   |                  |                  |
// 0      <=      readerIndex   <=   writerIndex    <=     size

type FixBuffer struct {
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

func readv(fd int, iovecs []syscall.Iovec, iovecLen int) (uintptr, syscall.Errno) {
	var (
		r uintptr
		e syscall.Errno
	)
	for {
		r, _, e = syscall.Syscall(syscall.SYS_READV, uintptr(fd), uintptr(unsafe.Pointer(&iovecs[0])), uintptr(iovecLen))
		if e != syscall.EINTR {
			break
		}
	}
	if e != 0 {
		return r, e
	}
	return r, 0
}

func (f *FixBuffer) ReadFd(fd int) (n int, syscallErr syscall.Errno) {
	var extraBuf [65536]byte
	var iovecsLen = 2
	iovecs := make([]syscall.Iovec, 2)

	writable := f.WritableBytes()
	iovecs[0].Base = &f.buf[f.wi]
	iovecs[0].Len = uint64(writable)
	iovecs[1].Base = &extraBuf[0]
	iovecs[1].Len = 65536

	if writable >= 65536 {
		iovecsLen = 1
	}
	r0, err := readv(fd, iovecs, iovecsLen)
	n = int(r0)
	if n < 0 {
		syscallErr = err
	} else if n <= writable {
		f.wi += n
	} else {
		f.wi = len(f.buf) // 移到最后
		f.Append(extraBuf[:n-writable])
	}
	return
}

func (f *FixBuffer) Swap(o *FixBuffer) {
	f.buf, o.buf = o.buf, f.buf
	f.ri, o.ri = o.ri, f.ri
	f.wi, o.wi = o.wi, f.wi
}

func (f *FixBuffer) ReadableBytes() int {
	return f.wi - f.ri
}

func (f *FixBuffer) WritableBytes() int {
	return len(f.buf) - f.wi
}

func (f *FixBuffer) PrependableBytes() int {
	return f.ri
}


// *************** retrieve *************** //
func (f *FixBuffer) Retrieve(len int) {
	if len > f.ReadableBytes() { // FIXME Error
		return
	}
	if len < f.ReadableBytes() {
		f.ri += len
	} else {
		f.RetrieveAll()
	}
}

func (f *FixBuffer) RetrieveAll() {
	f.ri, f.wi = cheapPrepend, cheapPrepend
}

func (f *FixBuffer) retrieveUint64() {
	f.Retrieve(8)
}

func (f *FixBuffer) retrieveUint32() {
	f.Retrieve(4)
}

func (f *FixBuffer) retrieveUint16() {
	f.Retrieve(2)
}

func (f *FixBuffer) retrieveUint8() {
	f.Retrieve(1)
}

func (f *FixBuffer) RetrieveAllAsString() string {
	ret := f.RetrieveAsBytes(f.ReadableBytes())
	return string(ret)
}

func (f *FixBuffer) RetrieveAsBytes(len int) (ret []byte) {
	if len <= f.ReadableBytes() {
		ret = f.Peek(len)
		f.Retrieve(len)
	}
	return
}



// ************** write / append **************** //
func (f *FixBuffer) Append(b []byte) {
	f.buf = append(f.buf[:len(f.buf)], b...)
}

func (f *FixBuffer) AppendByte(b byte) {
	f.buf = append(f.buf[:len(f.buf)], b)
}

func (f *FixBuffer) UnWrite(len int) {
	if len > f.ReadableBytes() {
		return
	}
	f.wi -= len
}

func (f *FixBuffer) HasWrite(len int) {
	if len <= f.WritableBytes() {
		f.wi += len
	}
}

func (f *FixBuffer) appendUint64(x uint64) {
	// 本机字节序 -> 网络字节序
	var xb []byte
	binary.BigEndian.PutUint64(xb, x)
	f.Append(xb)
	f.HasWrite(len(xb))
}

func (f *FixBuffer) appendUint32(x uint32) {
	// 本机字节序 -> 网络字节序
	var xb []byte
	binary.BigEndian.PutUint32(xb, x)
	f.Append(xb)
	f.HasWrite(len(xb))
}

func (f *FixBuffer) appendUint16(x uint16) {
	// 本机字节序 -> 网络字节序
	var xb []byte
	binary.BigEndian.PutUint16(xb, x)
	f.Append(xb)
	f.HasWrite(len(xb))
}

func (f *FixBuffer) appendUint8(x uint8) {
	f.AppendByte(x)
}

// ******************* read ******************* //
func (f *FixBuffer) ReadUint64() uint64 {
	x := f.PeekUint64()
	f.retrieveUint64()
	return x
}

func (f *FixBuffer) ReadUint32() uint32 {
	x := f.PeekUint32()
	f.retrieveUint32()
	return x
}

func (f *FixBuffer) ReadUint16() uint16 {
	x := f.PeekUint16()
	f.retrieveUint16()
	return x
}

func (f *FixBuffer) ReadUint8() uint8 {
	x := f.PeekUint8()
	f.retrieveUint8()
	return x
}

// ******************* peek ******************** //

func (f *FixBuffer) Peek(len int) (b []byte) {
	if len <= 0 {
		return
	}
	if len > f.ReadableBytes() { // log error
		return
	}
	b = f.buf[f.ri: f.ri+len]
	return
}

func (f *FixBuffer) PeekUint64() uint64 {
	b := f.Peek(8)
	if len(b) == 8 {
		return binary.BigEndian.Uint64(b)
	}
	return 0
}

func (f *FixBuffer) PeekUint32() uint32 {
	b := f.Peek(4)
	if len(b) == 4 {
		return binary.BigEndian.Uint32(b)
	}
	return 0
}

func (f *FixBuffer) PeekUint16() uint16 {
	b := f.Peek(4)
	if len(b) == 4 {
		return binary.BigEndian.Uint16(b)
	}
	return 0
}

func (f *FixBuffer) PeekUint8() uint8 {
	b := f.Peek(1)
	if len(b) == 1 {
		return b[0]
	}
	return 0

}


// ******************* prepend **************** //

func (f *FixBuffer) Prepend(b []byte) {
	if len(b) > f.PrependableBytes() {
		return
	}

	f.ri -= len(b)
	copy(f.buf[f.ri:], b)
	return
}

func (f *FixBuffer) PrependUint64(x uint64) {
	var xb []byte
	binary.BigEndian.PutUint64(xb, x)
	f.Prepend(xb)
}

func (f *FixBuffer) PrependUint32(x uint32) {
	var xb []byte
	binary.BigEndian.PutUint32(xb, x)
	f.Prepend(xb)
}

func (f *FixBuffer) PrependUint16(x uint16) {
	var xb []byte
	binary.BigEndian.PutUint16(xb, x)
	f.Prepend(xb)
}

func (f *FixBuffer) PrependUint8(x uint8) {
	f.Prepend([]byte{x})
}











