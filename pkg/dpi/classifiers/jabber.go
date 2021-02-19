package classifiers

import (
	"poor-man-service-mesh/pkg/dpi/types"
	"regexp"
)

// JABBERClassifier struct
type JABBERClassifier struct{}

// HeuristicClassify for JABBERClassifier
func (classifier JABBERClassifier) HeuristicClassify(packet *types.Packet) bool {
	payloadStr := string(packet.Payload)
	result, _ := regexp.MatchString("<?xml\\sversion='\\d+.\\d+'?.*", payloadStr)
	return result
}

// GetProtocol returns the corresponding protocol
func (classifier JABBERClassifier) GetProtocol() types.Protocol {
	return types.JABBER
}
