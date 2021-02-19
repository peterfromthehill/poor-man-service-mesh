package connection

import "net"

type Connection interface {
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
}
