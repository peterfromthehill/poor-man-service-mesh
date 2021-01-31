package connection

type Connection interface {
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
}
