package classifiers

import (
	"encoding/binary"
	"poor-man-service-mesh/pkg/dpi/types"

	"github.com/paultag/sniff/parser"
	"k8s.io/klog/v2"
)

// SSLClassifier struct
type SSLClassifier struct{}

// HeuristicClassify for SSLClassifier
func (classifier SSLClassifier) HeuristicClassify(packet *types.Packet) (detected bool) {
	payload := packet.Payload
	if len(payload) >= 9 {
		packetLen := int(binary.BigEndian.Uint16(payload[3:5]))
		clientHelloLenBytes := append([]byte{0}, payload[6:9]...)
		clientHelloLen := int(binary.BigEndian.Uint32(clientHelloLenBytes))
		// check if the packet looks like an SSL/TLS packet
		isSSLProto := payload[0] == 0x16 && payload[1] <= 0x03 && packetLen == len(payload[5:])
		// check if the first payload contains a ClientHello message
		isClientHello := payload[5] == 1 && clientHelloLen == len(payload[9:])
		if isClientHello {
			parseSNI(payload)
		}
		detected = isSSLProto && isClientHello
	}
	return
}

func parseSNI(payload []byte) {
	klog.Infof("%#x : %#x", payload[1:3], payload[9:11])
	extensions, err := parser.GetExtensionBlock(payload)
	if err != nil {
		klog.Infof("error parsing extensions: %s", err.Error())
		return
	}
	sn, err := parser.GetSNBlock(extensions)
	if err != nil {
		klog.Infof("error parsing SNI-Block: %s", err.Error())
		return
	}
	sni, err := parser.GetSNIBlock(sn)
	if err != nil {
		klog.Infof("error parsing SNI: %s", err.Error())
		return
	}
	klog.Infof("SNI: %s", string(sni))
}

// GetProtocol returns the corresponding protocol
func (classifier SSLClassifier) GetProtocol() types.Protocol {
	return types.SSL
}
