package classifiers

import (
	"poor-man-service-mesh/pkg/dpi/types"
	"strings"
)

// SMTPClassifier struct
type SMTPClassifier struct{}

// HeuristicClassify for SMTPClassifier
func (classifier SMTPClassifier) HeuristicClassify(packet *types.Packet) bool {
	payload := packet.Payload

	payloadStr := string(payload)
	for _, line := range strings.Split(payloadStr, "\n") {
		if len(line) > 0 && strings.HasPrefix(line, "220 ") {
			return true
		}
	}
	return (strings.HasPrefix(payloadStr, "EHLO ") ||
		strings.HasPrefix(payloadStr, "HELO ")) &&
		strings.HasSuffix(payloadStr, "\n")
}

// GetProtocol returns the corresponding protocol
func (classifier SMTPClassifier) GetProtocol() types.Protocol {
	return types.SMTP
}
