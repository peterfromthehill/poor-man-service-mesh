package types

import (
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

// Protocol is the type of each of the detected protocols.
type Protocol string

// Protocol identifiers for the supported protocols
const (
	HTTP       Protocol = "HTTP"
	DNS        Protocol = "DNS"
	SSH        Protocol = "SSH"
	RPC        Protocol = "RPC"
	SMTP       Protocol = "SMTP"
	RDP        Protocol = "RDP"
	SMB        Protocol = "SMB"
	ICMP       Protocol = "ICMP"
	FTP        Protocol = "FTP"
	SSL        Protocol = "SSL"
	NetBIOS    Protocol = "NetBIOS"
	JABBER     Protocol = "JABBER"
	MQTT       Protocol = "MQTT"
	BITTORRENT Protocol = "BitTorrent"
	Unknown    Protocol = ""
)

type Packet struct {
	Payload              []byte
	ClassificationResult []Protocol
}

func (p *Packet) AddClassificationResult(clazz Protocol) {
	p.ClassificationResult = append(p.ClassificationResult, clazz)
}

type Module interface {
	Initialize() error
	Destroy() error
	ClassifyFlow(packet *Packet) *Packet
}

var flowTracker *cache.Cache
var flowTrackerMtx sync.Mutex

// InitCache initializes the flow cache. It must be called before the cache
// is utilised. Flows will be discarded if they are inactive for the given
// duration. If that value is negative, flows will never expire.
func InitCache(expirationTime time.Duration) {
	flowTracker = cache.New(expirationTime, 5*time.Minute)
}

// DestroyCache frees the resources used by the flow cache.
func DestroyCache() {
	if flowTracker != nil {
		flowTracker.Flush()
		flowTracker = nil
	}
}
