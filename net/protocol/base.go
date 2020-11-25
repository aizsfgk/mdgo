package protocol


type Protocol interface {
	Pack()
	Unpack()
}