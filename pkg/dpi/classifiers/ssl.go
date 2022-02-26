package classifiers

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"io"
	"net"
	"poor-man-service-mesh/pkg/dpi/types"
	"time"

	"k8s.io/klog/v2"
)

// SSLClassifier struct
type SSLClassifier struct{}

// // HeuristicClassify for SSLClassifier
// func (classifier SSLClassifier) HeuristicClassify(packet *types.Packet) (detected bool) {
// 	payload := packet.Payload
// 	if len(payload) >= 9 {
// 		packetLen := int(binary.BigEndian.Uint16(payload[3:5]))
// 		clientHelloLenBytes := append([]byte{0}, payload[6:9]...)
// 		clientHelloLen := int(binary.BigEndian.Uint32(clientHelloLenBytes))
// 		// check if the packet looks like an SSL/TLS packet
// 		isSSLProto := payload[0] == 0x16 && payload[1] <= 0x03 && packetLen == len(payload[5:])
// 		// check if the first payload contains a ClientHello message
// 		isClientHello := payload[5] == 1 && clientHelloLen == len(payload[9:])
// 		if isClientHello {
// 			parseSNI(payload)
// 		}
// 		detected = isSSLProto && isClientHello
// 	}
// 	return
// }

type SNI struct {
	ServerName string
}

// HeuristicClassify for SSLClassifier
func (classifier SSLClassifier) HeuristicClassify(flow *types.Flow) (bool, interface{}) {
	sni := SNI{}
	return checkFirstPayload(flow.GetPackets(),
		func(payload []byte, _ []types.Packet) (detected bool) {
			if len(payload) >= 9 {
				packetLen := int(binary.BigEndian.Uint16(payload[3:5]))
				clientHelloLenBytes := append([]byte{0}, payload[6:9]...)
				clientHelloLen := int(binary.BigEndian.Uint32(clientHelloLenBytes))
				// check if the packet looks like an SSL/TLS packet
				isSSLProto := payload[0] == 22 && payload[1] <= 3 && packetLen == len(payload[5:])
				// check if the first payload contains a ClientHello message
				isClientHello := payload[5] == 1 && clientHelloLen == len(payload[9:])
				detected = isSSLProto && isClientHello
			}
			if detected {
				snireader := bytes.NewReader(payload)
				clientHello, err := readClientHello(snireader)
				if err != nil {
					klog.Warning(err)
				} else {
					klog.Warningf("ServerName: %s", clientHello.ServerName)
					sni.ServerName = clientHello.ServerName
				}

			}
			return
		}), sni
}

// func parseSNI(payload []byte) {
// 	klog.Infof("%#x : %#x", payload[1:3], payload[9:11])
// 	extensions, err := parser.GetExtensionBlock(payload)
// 	if err != nil {
// 		klog.Infof("error parsing extensions: %s", err.Error())
// 		return
// 	}
// 	sn, err := parser.GetSNBlock(extensions)
// 	if err != nil {
// 		klog.Infof("error parsing SNI-Block: %s", err.Error())
// 		return
// 	}
// 	sni, err := parser.GetSNIBlock(sn)
// 	if err != nil {
// 		klog.Infof("error parsing SNI: %s", err.Error())
// 		return
// 	}
// 	klog.Infof("SNI: %s", string(sni))
// }

// GetProtocol returns the corresponding protocol
func (classifier SSLClassifier) GetProtocol() types.Protocol {
	return types.SSL
}

func readClientHello(reader io.Reader) (*tls.ClientHelloInfo, error) {
	var hello *tls.ClientHelloInfo

	err := tls.Server(readOnlyConn{reader: reader}, &tls.Config{
		GetConfigForClient: func(argHello *tls.ClientHelloInfo) (*tls.Config, error) {
			hello = new(tls.ClientHelloInfo)
			*hello = *argHello
			return nil, nil
		},
	}).Handshake()

	if hello == nil {
		return nil, err
	}

	return hello, nil
}

type readOnlyConn struct {
	reader io.Reader
}

func (conn readOnlyConn) Read(p []byte) (int, error)         { return conn.reader.Read(p) }
func (conn readOnlyConn) Write(p []byte) (int, error)        { return 0, io.ErrClosedPipe }
func (conn readOnlyConn) Close() error                       { return nil }
func (conn readOnlyConn) LocalAddr() net.Addr                { return nil }
func (conn readOnlyConn) RemoteAddr() net.Addr               { return nil }
func (conn readOnlyConn) SetDeadline(t time.Time) error      { return nil }
func (conn readOnlyConn) SetReadDeadline(t time.Time) error  { return nil }
func (conn readOnlyConn) SetWriteDeadline(t time.Time) error { return nil }
