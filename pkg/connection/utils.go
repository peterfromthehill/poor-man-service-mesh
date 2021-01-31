package connection

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"reflect"
	"syscall"
	"unsafe"
)

// getConnFromTLSConn returns the internal wrapped connection from the tls.Conn.
func GetConnFromTLSConn(tlsConn *tls.Conn) net.Conn {
	// XXX: This is really BAD!!! Only way currently to get the underlying
	// connection of the tls.Conn. At least until
	// https://github.com/golang/go/issues/29257 is solved.
	conn := reflect.ValueOf(tlsConn).Elem().FieldByName("conn")
	conn = reflect.NewAt(conn.Type(), unsafe.Pointer(conn.UnsafeAddr())).Elem()
	return conn.Interface().(net.Conn)
}

func GetOrigAddr(conn Connection) (string, int, error) {
	file, err := getFileFromConn(conn)
	if err != nil {
		return "", 0, err
	}
	origAddr, origPort, err := getOrigAddrFromFile(file)
	if err != nil {
		return "", 0, err
	}
	return origAddr, origPort, nil
}

func getOrigAddrFromFile(file *os.File) (string, int, error) {
	const SO_ORIGINAL_DST = 80
	addr, err := syscall.GetsockoptIPv6Mreq(int(file.Fd()), syscall.IPPROTO_IP, SO_ORIGINAL_DST)
	if err != nil {
		return "", 0, fmt.Errorf("syscall.GetsockoptIPv6Mreq error: %w", err)
	}

	remoteAddr := fmt.Sprintf("%d.%d.%d.%d",
		addr.Multiaddr[4], addr.Multiaddr[5], addr.Multiaddr[6], addr.Multiaddr[7])
	remotePort := int(uint16(addr.Multiaddr[2])<<8 + uint16(addr.Multiaddr[3]))
	return remoteAddr, remotePort, nil
}

func getFileFromConn(netConn Connection) (*os.File, error) {
	if tcpConn, ok := netConn.(*net.TCPConn); ok {
		file, err := getFileFromTCPConn(tcpConn)
		if err != nil {
			return nil, err
		}
		return file, nil
	}
	return nil, errors.New("not a tcp or tls connection")
}

func getFileFromTCPConn(tcpConn *net.TCPConn) (*os.File, error) {
	file, err := tcpConn.File()
	if err != nil {
		return nil, err
	}
	return file, nil
}
