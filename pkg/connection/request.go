package connection

import (
	"crypto/tls"
	"errors"
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
	protocol  map[types.Protocol]types.ClassificationResult
}

func NewRequestWithDst(srcConn Connection, router string, origAddr string, origPort int, secure bool) (*Request, error) {
	request := &Request{
		srcClient: srcConn,
		dstClient: nil,
		dstHost:   origAddr,
		dstPort:   origPort,
		router:    router,
		wg:        &sync.WaitGroup{},
		protocol:  make(map[types.Protocol]types.ClassificationResult),
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
		return nil, fmt.Errorf("can not get real destination from conntrack: %s", err.Error())
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

/*
  Create the "tunneled" connection through a HTTP/Connect connection.
*/
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

/*
  Create a HTTP/Connect "tunnel" connection.
  Should the a go transport/dialer?!
*/
func (request *Request) connectToDestination() error {
	strHostPort := fmt.Sprintf("%s:%d", request.dstHost, request.dstPort)
	strEcho := "CONNECT " + strHostPort + " HTTP/1.1\n" +
		"Host: " + request.dstHost + "\n" +
		"Proxy-Connection: Keep-Alive\n\n"
	_, err := request.dstClient.Write([]byte(strEcho))
	return err
}

// func (request *Request) readUntilHeaderEnd(reader *bufio.Reader) ([]byte, error) {
// 	// header := []byte{'\n', '\r', '\n', '\r'}
// 	// slice := []byte{0x00, 0x00, 0x00, 0x00}
// 	// fmt.Println(slice)
// 	// singleByte := make([]byte, 1)
// 	// for {
// 	// 	request.dstClient.Read(singleByte)
// 	// 	slice = slice[1:]
// 	// 	slice = append(slice, singleByte...)
// 	// 	fmt.Println(string(slice))
// 	// 	if equal(header, slice) {
// 	// 		return nil, nil
// 	// 	}
// 	// }
// 	http.ReadRequest(reader)
// 	return nil, nil
// }

/*
  Read max 1024 byte from the remote proxy answer with a HTTP status code.
  TODO: overreading bytes should be added to a multireader
  https://www.geeksforgeeks.org/io-multireader-function-in-golang-with-examples/
*/
func (request *Request) readUntilReady() error {
	reply := make([]byte, 1024)
	//reply, _ := request.readUntilHeaderEnd(request.dstClient)

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
	return fmt.Errorf("failed Connect to Host (%s:%d): %d", request.dstHost, request.dstPort, statusCode)
}

func (request *Request) transfer(source Connection, dest Connection, dataChan chan []byte) error {
	//defer source.Close()
	//defer dest.Close()

	var err error
	for {
		readBuffer := make([]byte, 1024)
		readLen, err := source.Read(readBuffer)
		if err != nil || readLen == 0 {
			//klog.Warningf("Error readin")
			return err
		} else {
			if request.dstClient != nil {
				sendBuffer := make([]byte, readLen)
				copy(sendBuffer, readBuffer)
				dataChan <- sendBuffer
				nw, err := dest.Write(sendBuffer)
				if nw < 0 || readLen < nw {
					nw = 0
					if err == nil {
						err = errors.New("invalid write result")
					}
				}
				if err != nil {
					break
				}
				if readLen != nw {
					err = errors.New("short write")
					break
				}
			}
		}
	}
	return err
	//klog.Warningf("Error reading/writing, close connections")
}

func (request *Request) dpi(ctrl chan bool, packetChan chan []byte) {
	// DPI
	// funktioniert so nicht, weil es bekommt nur die daten von einem Datenstrom.
	// Idee: gofunc starten mit einem channel reader und beide Streams schreiben in diesen Channel
	dpi.Initialize()
	dataflow := types.NewFlow()
	defer dpi.Destroy()
	// DPI END
	for {
		select {
		case <-ctrl:
			resultProtos := dpi.ClassifyFlow(dataflow)
			if len(resultProtos) > 0 {
				klog.Infof("%q", resultProtos)
			}
			for _, v := range resultProtos {
				request.protocol[v.Protocol] = v
			}
			ctrl <- true
			break
		case analyzeBuffer := <-packetChan:
			packet := dpi.GetPacket(analyzeBuffer)
			dataflow.AddPacket(packet)
			klog.Infof("%d Pakets in Flow", len(dataflow.GetPackets()))
		}
	}
}

func (request *Request) Listen() {
	ctrl := make(chan bool)
	dataChan := make(chan []byte)
	go request.dpi(ctrl, dataChan)
	defer func() {
		request.wg.Wait()
		ctrl <- false
		<-ctrl
	}()
	request.wg.Add(2)
	go func() {
		defer request.wg.Done()
		err := request.transfer(request.srcClient, request.dstClient, dataChan)
		_ = err
		//klog.Errorf("Connection closed: %s", err)
		request.srcClient.(*net.TCPConn).CloseWrite()
		request.dstClient.(*net.TCPConn).CloseWrite()

	}()
	go func() {
		defer request.wg.Done()
		err := request.transfer(request.dstClient, request.srcClient, dataChan)
		_ = err
		//klog.Errorf("Connection closed: %s", err)
		request.dstClient.(*net.TCPConn).CloseWrite()
		request.srcClient.(*net.TCPConn).CloseWrite()
	}()
}

func (request *Request) String() string {
	protocolStr := ""

	for _, j := range request.protocol {
		protocolStr = fmt.Sprintf("%s %s: %s", protocolStr, j.Protocol, j.Details)
	}

	return fmt.Sprintf("%s -> %s (%s)", request.srcClient.RemoteAddr(), request.dstClient.RemoteAddr(), protocolStr)
}
