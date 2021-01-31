package eproxy

import (
	"io"
	"net"
	"net/http"
	"poor-man-service-mesh/pkg/connection"
	"poor-man-service-mesh/pkg/proxyserver"
	"strconv"

	"k8s.io/klog/v2"
)

type Eproxy struct {
	// port            int
	// sslKeyFilePath  string
	// sslCertFilePath string
	// directoryURL    string
	// secure          bool
	externalProxy string
}

func Build(externalProxy string) proxyserver.HttpProxyInstance {
	klog.Infof("Initilize eProxy with PROXY: %s", externalProxy)
	return &Eproxy{
		externalProxy: externalProxy,
	}
}

func (eproxy *Eproxy) findFQDNs() ([]string, error) {
	externalIP, err := externalIP()
	if err != nil {
		klog.Error(err.Error())
		return nil, err
	}
	externalIPinAddr, _ := reverseaddr(externalIP)
	rDNSNames, err := reverseLookup(externalIPinAddr)
	if err != nil {
		klog.Error(err.Error())
		return nil, err
	}
	return rDNSNames, nil
}

func (eproxy *Eproxy) HandleTunneling(w http.ResponseWriter, r *http.Request) {
	// host, port, err := net.SplitHostPort(r.Host)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusServiceUnavailable)
	// 	return
	// }
	// dstHost := net.JoinHostPort(host, port)
	// fmt.Printf("Incomming request for %s -> rewrite to: %s\n", r.Host, dstHost)
	klog.Infof("verbinde zu %s über %s", r.Host, eproxy.externalProxy)
	klog.Infof("%s", r)

	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		klog.Infof("Connection timeout\n")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	hostPort := r.Host
	if hostPort == "" {
		hostPort = r.URL.Host
	}

	host, sPort, err := net.SplitHostPort(hostPort)
	if err != nil {
		klog.Infof("%s? %s", hostPort, err.Error())
	}
	port, err := strconv.Atoi(sPort)
	if err != nil {
		klog.Infof("%s? %s", sPort, err.Error())
	}

	_, err = connection.NewRequestWithDst(clientConn, eproxy.externalProxy, host, port, false)
	if err != nil {
		klog.Infof("Request failed: %w", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	// destConn, err := net.DialTimeout("tcp", eproxy.externalProxy, 10*time.Second)
	// if err != nil {
	// 	fmt.Printf("Connection timeout\n")
	// 	http.Error(w, err.Error(), http.StatusServiceUnavailable)
	// 	return
	// }

	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusServiceUnavailable)
	// }
	// go eproxy.debugTransfer(destConn, clientConn)
	// go eproxy.debugTransfer(clientConn, destConn)
}

func (eproxy *Eproxy) transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func (eproxy *Eproxy) debugTransfer(source net.Conn, dest net.Conn) {
	defer source.Close()
	defer dest.Close()
	for {
		readBuffer := make([]byte, 1024)
		readLen, err := source.Read(readBuffer)
		if err != nil {
			klog.Warningf("Error readin")
			break
		} else {
			klog.Infof("Recv byte len: %d", readLen)
			sendBuffer := make([]byte, readLen)
			copy(sendBuffer, readBuffer)
			klog.Infof("> %s", string(sendBuffer))
			_, err := dest.Write(sendBuffer)
			if err != nil {
				klog.Warningf("Error writing")
				break
			}
		}
	}
	klog.Warningf("Error reading/writing, close connections")
}
