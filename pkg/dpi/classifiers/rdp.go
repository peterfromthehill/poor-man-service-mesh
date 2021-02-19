package classifiers

import (
	"encoding/binary"
	"poor-man-service-mesh/pkg/dpi/types"
	"strings"
)

// RDPClassifier struct
type RDPClassifier struct{}

// HeuristicClassify for RDPClassifier
func (classifier RDPClassifier) HeuristicClassify(packet *types.Packet) bool {
	payload := packet.Payload
	if len(payload) < 20 {
		return false
	}
	tpktLen := int(binary.BigEndian.Uint16(payload[2:4]))
	// check TPKT header
	isValidTpkt := payload[0] == 3 && payload[1] == 0 && tpktLen == len(payload)
	// check COTP header
	isValidCotp := int(payload[4]) == len(payload[5:]) && payload[5] == 0xE0
	// check RDP payload
	rdpPayloadStr := string(payload[11:])
	isValidRdp := strings.Contains(rdpPayloadStr, "mstshash=") ||
		strings.Contains(rdpPayloadStr, "msts=")
	return isValidTpkt && isValidCotp && isValidRdp
}

// GetProtocol returns the corresponding protocol
func (classifier RDPClassifier) GetProtocol() types.Protocol {
	return types.RDP
}
