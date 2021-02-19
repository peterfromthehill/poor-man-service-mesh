package tproxy

import (
	"fmt"
	"net"
	"poor-man-service-mesh/pkg/connection"
	"strings"

	"k8s.io/klog"
)

type TProxy struct {
	port        int
	serviceCidr string
	podCidr     string
	exitProxy   string
	secure      bool
}

func NewTProxy(port int, serviceCidr string, podCidr string, exitProxy string, secure bool) *TProxy {
	return &TProxy{
		port:        port,
		serviceCidr: serviceCidr,
		podCidr:     podCidr,
		exitProxy:   exitProxy,
		secure:      secure,
	}
}

func (tproxy *TProxy) Listen() error {
	klog.Infof("Waiting for outbound connections on port %d", tproxy.port)
	listener, err := net.Listen("tcp", ":"+fmt.Sprintf("%d", tproxy.port))
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		go tproxy.routeRequest(conn)
	}
}

func (tproxy *TProxy) routeRequest(conn connection.Connection) {
	origAddr, origPort, err := connection.GetOrigAddr(conn)
	if err != nil {
		klog.Warning("Error with connection")
		return
	}
	origAddrIP := net.ParseIP(origAddr)

	routerDNS, err := net.LookupAddr(origAddr)
	if err != nil {
		klog.Warningf("Cannot resolve IP: %s -> use IP instand", origAddr)
		// Maybe always?
		routerDNS = []string{origAddr}
	}
	klog.Infof("Resolved to hostname: %s", routerDNS)

	var request *connection.Request
	if tproxy.isIPInside(origAddrIP) {
		klog.Infof("Route internal: %s:%d", origAddrIP, origPort)
		dstrouter := fmt.Sprintf("%s:%d", strings.Trim(routerDNS[0], "."), origPort)
		request, err = connection.NewRequest(conn, dstrouter, tproxy.secure)
	} else {
		klog.Infof("Route external: %s", origAddrIP)
		request, err = connection.NewRequest(conn, tproxy.exitProxy, tproxy.secure)
	}
	if err != nil {
		klog.Errorf(err.Error())
	}
	klog.Infof("Listen!")
	request.Listen()
	klog.Infof("after Listen!")
	klog.Infof(request.String())
}

func (tproxy *TProxy) isIPInside(ip net.IP) bool {
	_, serviceNet, err := net.ParseCIDR(tproxy.serviceCidr)
	if err != nil {
		klog.Errorf(err.Error())
	}

	_, podNet, err := net.ParseCIDR(tproxy.podCidr)
	if err != nil {
		klog.Errorf(err.Error())
	}

	if serviceNet.Contains(ip) || podNet.Contains(ip) {
		return true
	}
	return false
}
