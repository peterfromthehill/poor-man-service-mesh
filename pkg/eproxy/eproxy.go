package eproxy

import (
	"net"
	"net/http"
	"poor-man-service-mesh/pkg/connection"
	"poor-man-service-mesh/pkg/proxyserver"
	"strconv"

	"k8s.io/klog/v2"
)

type Eproxy struct {
	externalProxy string
}

func Build(externalProxy string) proxyserver.HttpProxyInstance {
	klog.Infof("Initilize eProxy with PROXY: %s", externalProxy)
	return &Eproxy{
		externalProxy: externalProxy,
	}
}

func (eproxy *Eproxy) HandleTunneling(w http.ResponseWriter, r *http.Request) {
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
		klog.Errorf("%s: No host/port information - you are in an mesh? Err: %s", hostPort, err.Error())
		return
	}
	port, err := strconv.Atoi(sPort)
	if err != nil {
		klog.Errorf("%s is not a number", sPort, err.Error())
		return
	}

	_, err = connection.NewRequestWithDst(clientConn, eproxy.externalProxy, host, port, false)
	if err != nil {
		klog.Infof("Request failed: %w", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
}
