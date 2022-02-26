package classifiers

import (
	"encoding/binary"
	"poor-man-service-mesh/pkg/dpi/types"
	"strings"
)

// SMBClassifier struct
type SMBClassifier struct{}

// // HeuristicClassify for SMBClassifier
// func (classifier SMBClassifier) HeuristicClassify(packet *types.Packet) bool {
// 	payload := packet.Payload
// 	// skip netbios layer if it exists
// 	if len(payload) > 4 && payload[0] == 0 {
// 		netbiosLen := binary.BigEndian.Uint32(payload[:4])
// 		if int(netbiosLen) == len(payload[4:]) {
// 			payload = payload[4:]
// 		}
// 	}
// 	if len(payload) < 10 {
// 		return false
// 	}
// 	// SMB protocol prefix
// 	hasSMBPrefix := strings.HasPrefix(string(payload), "\xFFSMB")
// 	// SMB protocol negotiation code
// 	isNegotiateProtocol := payload[4] == 0x72
// 	// error code must be zero
// 	errCode := binary.BigEndian.Uint32(payload[5:9])
// 	// if flag is 0 this packet is from the server to the client
// 	directionFlag := payload[9] & 0x80
// 	return hasSMBPrefix && isNegotiateProtocol && errCode == 0 && directionFlag == 0

// }

// HeuristicClassify for SMBClassifier
func (classifier SMBClassifier) HeuristicClassify(flow *types.Flow) (bool, interface{}) {
	return checkFirstPayload(flow.GetPackets(),
		func(payload []byte, packetsRest []types.Packet) bool {
			// skip netbios layer if it exists
			if len(payload) > 4 && payload[0] == 0 {
				netbiosLen := binary.BigEndian.Uint32(payload[:4])
				if int(netbiosLen) == len(payload[4:]) {
					payload = payload[4:]
				}
			}
			if len(payload) < 10 {
				return false
			}
			// SMB protocol prefix
			hasSMBPrefix := strings.HasPrefix(string(payload), "\xFFSMB")
			// SMB protocol negotiation code
			isNegotiateProtocol := payload[4] == 0x72
			// error code must be zero
			errCode := binary.BigEndian.Uint32(payload[5:9])
			// if flag is 0 this packet is from the server to the client
			directionFlag := payload[9] & 0x80
			return hasSMBPrefix && isNegotiateProtocol && errCode == 0 && directionFlag == 0
		}), struct{}{}
}

// GetProtocol returns the corresponding protocol
func (classifier SMBClassifier) GetProtocol() types.Protocol {
	return types.SMB
}
