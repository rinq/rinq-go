package overpass

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// PeerID uniquely identifies a peer within a network.
// Peer IDs are designed to be somewhat human-readable.
type PeerID struct {
	// Clock is a time-based portion of the ID, this helps uniquely identify
	// peer IDs over longer time-scales, such as when looking back through
	// logs, etc.
	Clock uint64

	// Rand is a unique number identifying this peer within a network at any
	// given time. It is generated randomly and then reserved when the peer
	// connects to the network.
	Rand uint16
}

// NewPeerID creates a new ID struct. There is no guarantee that the ID is
// unique until the peer is registered with the network.
func NewPeerID() PeerID {
	return PeerID{
		uint64(time.Now().Unix()),
		uint16(rand.Intn(math.MaxUint16-1)) + 1,
	}
}

// Validate returns nil if the ID is valid.
func (id PeerID) Validate() error {
	if id.Clock != 0 && id.Rand != 0 {
		return nil
	}

	return fmt.Errorf("Peer ID %s is invalid.", id)
}

// ShortString returns a string representation of the "Rand" component.
func (id PeerID) ShortString() string {
	return fmt.Sprintf(
		"%04X",
		id.Rand,
	)
}

func (id PeerID) String() string {
	return fmt.Sprintf(
		"%X-%s",
		id.Clock,
		id.ShortString(),
	)
}
