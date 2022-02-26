package types

import (
	"sync"
)

type ClassificationSource string

type ClassificationResult struct {
	Protocol Protocol
	Details  interface{}
}

type Flow struct {
	packets []Packet
	//classification       ClassificationResult
	ClassificationResult []ClassificationResult
	mtx                  sync.RWMutex
}

func (flow *Flow) AddClassificationResult(clazz ClassificationResult) {
	flow.ClassificationResult = append(flow.ClassificationResult, clazz)
}

// NewFlow creates an empty flow.
func NewFlow() (flow *Flow) {
	flow = new(Flow)
	flow.packets = make([]Packet, 0)
	return
}

// CreateFlowFromPacket creates a flow with a single packet.
func CreateFlowFromPacket(packet Packet) (flow *Flow) {
	flow = NewFlow()
	flow.AddPacket(packet)
	return
}

// AddPacket adds a new packet to the flow.
func (flow *Flow) AddPacket(packet Packet) {
	flow.mtx.Lock()
	flow.packets = append(flow.packets, packet)
	flow.mtx.Unlock()
}

// GetPackets returns the list of packets in a thread-safe way.
func (flow *Flow) GetPackets() (packets []Packet) {
	flow.mtx.RLock()
	packets = make([]Packet, len(flow.packets))
	copy(packets, flow.packets)
	flow.mtx.RUnlock()
	return
}
