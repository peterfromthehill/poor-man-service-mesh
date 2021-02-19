package connection

import (
	"crypto/tls"
	"fmt"
	"net"
	"poor-man-service-mesh/pkg/certpool"
	"poor-man-service-mesh/pkg/dpi"
	"poor-man-service-mesh/pkg/dpi/types"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

type Request struct {
	srcClient Connection
	dstClient Connection
	dstHost   string
	dstPort   int
	router    string
	wg        *sync.WaitGroup
	protocol  map[types.Protocol]struct{}
}

func NewRequestWithDst(srcConn Connection, router string, origAddr string, origPort int, secure bool) (*Request, error) {
	request := &Request{
		srcClient: srcConn,
		dstClient: nil,
		dstHost:   origAddr,
		dstPort:   origPort,
		router:    router,
		wg:        &sync.WaitGroup{},
		protocol:  make(map[types.Protocol]struct{}),
	}
	var err error
	if secure {
		err = request.establishSecureConnectionThroughRouter()
	} else {
		err = request.establishInsecureConnectionThroughRouter()
	}
	if err != nil {
		return nil, err
	}
	return request, nil
}

func NewRequest(srcConn Connection, router string, secure bool) (*Request, error) {
	origAddr, origPort, err := GetOrigAddr(srcConn)
	if err != nil {
		return nil, fmt.Errorf("Can not get real destination from conntrack: %s", err.Error())
	}
	request, err := NewRequestWithDst(srcConn, router, origAddr, origPort, secure)
	return request, err
}

func (request *Request) establishInsecureConnectionThroughRouter() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", request.router)
	if err != nil {
		klog.Warningf("ResolveTCPAddr failed: %s", err.Error())
		return err
	}

	dstConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		klog.Errorf("Dial failed: %s", err.Error())
		return err
	}

	dstClient := dstConn
	request.dstClient = dstClient
	err = request.handshake()
	if err != nil {
		return err
	}
	return nil
}

func (request *Request) establishSecureConnectionThroughRouter() error {
	conf := &tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            certpool.GetCache().RootCAs,
	}

	klog.Infof("connecting to %s", request.router)
	dstConn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: time.Second * 2},
		"tcp",
		request.router,
		conf)
	if err != nil {
		return err
	}

	dstClient := dstConn
	request.dstClient = dstClient
	err = request.handshake()
	if err != nil {
		return err
	}
	return nil

}
func (request *Request) handshake() error {
	err := request.connectToDestination()
	if err != nil {
		return err
	}

	err = request.readUntilReady()
	if err != nil {
		return err
	}
	return nil
}

func (request *Request) connectToDestination() error {
	strHostPort := fmt.Sprintf("%s:%d", request.dstHost, request.dstPort)
	strEcho := "CONNECT " + strHostPort + " HTTP/1.1\n" +
		"Host: " + request.dstHost + "\n" +
		"Proxy-Connection: Keep-Alive\n\n"
	_, err := request.dstClient.Write([]byte(strEcho))
	return err
}

func (request *Request) readUntilReady() error {
	reply := make([]byte, 1024)
	replyLen, err := request.dstClient.Read(reply)
	if err != nil {
		klog.Warningf("Read to server failed:", err.Error())
		return err
	}
	_ = replyLen
	//klog.Infof("Revc byte len: %d", replyLen)
	statusCode, err := strconv.Atoi(strings.Split(string(reply), " ")[1])
	if err != nil {
		klog.Errorf(err.Error())
		return err
	}
	if statusCode >= 200 || statusCode <= 300 {
		klog.Infof("Connection to real destination (%s:%d) successful", request.dstHost, request.dstPort)
		return nil
	}
	return fmt.Errorf("Failed Connect to Host (%s:%d): %d", request.dstHost, request.dstPort, statusCode)
}

func (request *Request) transfer(source Connection, dest Connection) {
	defer source.Close()
	defer dest.Close()
	defer request.wg.Done()
	dpi.Initialize()
	defer dpi.Destroy()
	for {
		readBuffer := make([]byte, 1024)
		readLen, err := source.Read(readBuffer)
		if err != nil {
			//klog.Warningf("Error readin")
			break
		} else {
			analyzeBuffer := make([]byte, readLen)
			copy(analyzeBuffer, readBuffer)

			packet := dpi.GetPacket(analyzeBuffer)
			resultProtos := dpi.ClassifyFlow(packet)
			klog.Infof("%q", resultProtos)
			for _, v := range resultProtos {
				request.protocol[v] = struct{}{}
			}
			if request.dstClient != nil {
				sendBuffer := make([]byte, readLen)
				copy(sendBuffer, readBuffer)
				_, err := dest.Write(sendBuffer)
				if err != nil {
					//klog.Warningf("Error writing")
					break
				}
			}
		}
	}
	//klog.Warningf("Error reading/writing, close connections")
}

func (request *Request) Listen() {
	defer request.wg.Wait()
	request.wg.Add(1)
	go request.transfer(request.srcClient, request.dstClient)
	request.wg.Add(1)
	go request.transfer(request.dstClient, request.srcClient)
}

func (request *Request) String() string {
	return fmt.Sprintf("%s -> %s %q", request.srcClient.RemoteAddr(), request.dstClient.RemoteAddr(), func() []types.Protocol {
		var ret []types.Protocol
		for k, _ := range request.protocol {
			ret = append(ret, k)
		}
		return ret
	}())
}
