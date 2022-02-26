package classifiers

import (
	"poor-man-service-mesh/pkg/dpi/types"
	"strings"
)

// SMTPClassifier struct
type SMTPClassifier struct{}

// HeuristicClassify for SMTPClassifier
func (classifier SMTPClassifier) HeuristicClassify(flow *types.Flow) (bool, interface{}) {
	return checkFirstPayload(flow.GetPackets(),
		func(payload []byte, packetsRest []types.Packet) bool {
			payloadStr := string(payload)
			for _, line := range strings.Split(payloadStr, "\n") {
				if len(line) > 0 && !strings.HasPrefix(line, "220") {
					return false
				}
			}
			return checkFirstPayload(packetsRest,
				func(payload []byte, _ []types.Packet) bool {
					payloadStr := string(payload)
					return (strings.HasPrefix(payloadStr, "EHLO ") ||
						strings.HasPrefix(payloadStr, "HELO ")) &&
						strings.HasSuffix(payloadStr, "\n")
				})
		}), struct{}{}
}

// GetProtocol returns the corresponding protocol
func (classifier SMTPClassifier) GetProtocol() types.Protocol {
	return types.SMTP
}
