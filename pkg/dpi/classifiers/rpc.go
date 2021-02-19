package classifiers

import (
	"bytes"
	"encoding/binary"
	"poor-man-service-mesh/pkg/dpi/types"
)

// RPCClassifier struct
type RPCClassifier struct{}

// HeuristicClassify for RPCClassifier
func (classifier RPCClassifier) HeuristicClassify(packet *types.Packet) bool {
	payload := packet.Payload
	if len(payload) < 24 {
		return false
	}
	// check first bytes for version 5.0 bind request
	firstBytes := []byte{5, 0, 11, 3, 16, 0, 0, 0}
	// check if lengths match
	frameLen := int(binary.LittleEndian.Uint16(payload[8:10]))
	return bytes.HasPrefix(payload, firstBytes) && len(payload) == frameLen
}

// GetProtocol returns the corresponding protocol
func (classifier RPCClassifier) GetProtocol() types.Protocol {
	return types.RPC
}
