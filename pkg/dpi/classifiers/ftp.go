package classifiers

import (
	"poor-man-service-mesh/pkg/dpi/types"
	"strings"
)

// FTPClassifier struct
type FTPClassifier struct{}

// // HeuristicClassify for FTPClassifier
// func (classifier FTPClassifier) HeuristicClassify(flow *types.Flow) bool {
// 	return checkFirstPayload(flow.GetPackets(), layers.LayerTypeTCP,
// 		func(payload []byte, packetsRest []gopacket.Packet) bool {
// 			payloadStr := string(payload)
// 			for _, line := range strings.Split(payloadStr, "\n") {
// 				if len(line) > 0 && !strings.HasPrefix(line, "220") {
// 					return false
// 				}
// 			}
// 			return checkFirstPayload(packetsRest, layers.LayerTypeTCP,
// 				func(payload []byte, _ []gopacket.Packet) bool {
// 					payloadStr := string(payload)
// 					return strings.HasPrefix(payloadStr, "USER ") &&
// 						strings.HasSuffix(payloadStr, "\n")
// 				})
// 		})
// }

// HeuristicClassify for FTPClassifier
func (classifier FTPClassifier) HeuristicClassify(packet *types.Packet) bool {
	payloadStr := string(packet.Payload)
	for _, line := range strings.Split(payloadStr, "\n") {
		if len(line) > 0 && !strings.HasPrefix(line, "220") {
			return false
		}
	}
	return strings.HasPrefix(payloadStr, "USER ") &&
		strings.HasSuffix(payloadStr, "\n")

}

// GetProtocol returns the corresponding protocol
func (classifier FTPClassifier) GetProtocol() types.Protocol {
	return types.FTP
}
