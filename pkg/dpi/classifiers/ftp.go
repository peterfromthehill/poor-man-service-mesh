package classifiers

import (
	"poor-man-service-mesh/pkg/dpi/types"
	"strings"
)

type FTP struct {
	User string
	Pass string
}

// FTPClassifier struct
type FTPClassifier struct{}

// // HeuristicClassify for FTPClassifier
func (classifier FTPClassifier) HeuristicClassify(flow *types.Flow) (bool, interface{}) {
	ftp := FTP{}
	return checkFirstPayload(flow.GetPackets(),
		func(payload []byte, packetsRest []types.Packet) bool {
			payloadStr := string(payload)
			payloadStr = strings.ReplaceAll(payloadStr, "\r", "")
			for _, line := range strings.Split(payloadStr, "\n") {
				if len(line) > 0 && !strings.HasPrefix(line, "220") {
					return false
				}
			}
			return checkFirstPayload(packetsRest,
				func(payload []byte, _ []types.Packet) bool {
					payloadStr := string(payload)
					isFTP := strings.HasPrefix(payloadStr, "USER ") && strings.HasSuffix(payloadStr, "\n")
					if isFTP {
						for _, p := range flow.GetPackets() {
							payloadStr := string(p.Payload)
							payloadStr = strings.ReplaceAll(payloadStr, "\r", "")
							for _, line := range strings.Split(payloadStr, "\n") {
								if strings.HasPrefix(line, "USER ") {
									ftp.User = strings.Split(line, " ")[1]
								}
								if strings.HasPrefix(line, "PASS ") {
									ftp.Pass = strings.Split(line, " ")[1]
								}
							}
						}
					}
					return isFTP
				})
		}), ftp
}

// HeuristicClassify for FTPClassifier
// func (classifier FTPClassifier) HeuristicClassify(packet *types.Packet) bool {
// 	payloadStr := string(packet.Payload)
// 	//klog.Warningf("%s", payloadStr)
// 	for _, line := range strings.Split(payloadStr, "\n") {
// 		// Lines have to start with "220"
// 		if len(line) > 0 && !strings.HasPrefix(line, "220") {
// 			return false
// 		}
// 	}
// 	return strings.HasPrefix(payloadStr, "USER ") &&
// 		strings.HasSuffix(payloadStr, "\n")

// }

// GetProtocol returns the corresponding protocol
func (classifier FTPClassifier) GetProtocol() types.Protocol {
	return types.FTP
}
