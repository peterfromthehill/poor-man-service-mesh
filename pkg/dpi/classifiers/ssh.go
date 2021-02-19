package classifiers

import (
	"poor-man-service-mesh/pkg/dpi/types"
	"strings"
)

// SSHClassifier struct
type SSHClassifier struct{}

// HeuristicClassify for SSHClassifier
func (classifier SSHClassifier) HeuristicClassify(packet *types.Packet) bool {
	payload := packet.Payload
	payloadStr := string(payload)
	hasSuffix := strings.HasSuffix(payloadStr, "\n")
	hasSSHStr := strings.HasPrefix(payloadStr, "SSH") || strings.Contains(payloadStr, "OpenSSH")
	return hasSuffix && hasSSHStr
}

// GetProtocol returns the corresponding protocol
func (classifier SSHClassifier) GetProtocol() types.Protocol {
	return types.SSH
}
