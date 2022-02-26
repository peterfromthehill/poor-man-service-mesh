package classifiers

import (
	"poor-man-service-mesh/pkg/dpi/types"
	"regexp"
)

// JABBERClassifier struct
type JABBERClassifier struct{}

// HeuristicClassify for JABBERClassifier
func (classifier JABBERClassifier) HeuristicClassify(flow *types.Flow) (bool, interface{}) {
	return checkFirstPayload(flow.GetPackets(),
		func(payload []byte, packetsRest []types.Packet) bool {
			payloadStr := string(payload)
			result, _ := regexp.MatchString("<?xml\\sversion='\\d+.\\d+'?.*", payloadStr)
			return result
		}), struct{}{}
}

// GetProtocol returns the corresponding protocol
func (classifier JABBERClassifier) GetProtocol() types.Protocol {
	return types.JABBER
}
