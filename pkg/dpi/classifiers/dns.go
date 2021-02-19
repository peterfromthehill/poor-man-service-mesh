package classifiers

import (
	"poor-man-service-mesh/pkg/dpi/types"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// DNSClassifier struct
type DNSClassifier struct{}

// HeuristicClassify for DNSClassifier
func (classifier DNSClassifier) HeuristicClassify(packet *types.Packet) (detected bool) {
	defer func() {
		if err := recover(); err != nil {
			detected = false
		}
	}()
	layerParser := gopacket.DecodingLayerParser{}
	dns := layers.DNS{}
	err := dns.DecodeFromBytes(packet.Payload, &layerParser)
	// attempt to decode layer as DNS packet using gopacket and return
	// whether it was successful
	detected = err == nil
	return
}

// GetProtocol returns the corresponding protocol
func (classifier DNSClassifier) GetProtocol() types.Protocol {
	return types.DNS
}
