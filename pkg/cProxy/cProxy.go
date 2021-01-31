package cProxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"poor-man-service-mesh/pkg/proxyserver"

	//"poor-man-service-mesh/pkg/autocert"

	"time"
)

const (
	LOCALHOST = "127.0.0.1"
)

type CProxy struct {
	// port            int
	// sslKeyFilePath  string
	// sslCertFilePath string
	// directoryURL    string
	// secure          bool
}

func Build() proxyserver.HttpProxyInstance {
	return &CProxy{}
}

func (cProxy *CProxy) HandleTunneling(w http.ResponseWriter, r *http.Request) {
	_, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	dstHost := net.JoinHostPort(LOCALHOST, port)
	fmt.Printf("Incomming request for %s -> rewrite to: %s\n", r.Host, dstHost)
	destConn, err := net.DialTimeout("tcp", dstHost, 10*time.Second)
	if err != nil {
		fmt.Printf("Connection timeout\n")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go cProxy.transfer(destConn, clientConn)
	go cProxy.transfer(clientConn, destConn)
}

func (cProxy *CProxy) transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}
