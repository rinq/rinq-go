package internals

import (
	"context"

	"github.com/over-pass/overpass-go/src/overpass"
)

// Invoker is a low-level RPC interface, it is used to implement the
// "command subsystem", as well as internal peer-to-peer requests.
//
// The terminology "call" refers to an invocation that expects a response,
// whereas "execute" is an invocation where no response is required.
type Invoker interface {
	Service

	// CallUnicast sends a unicast command request to a specific peer and blocks
	// until a response is received or the context deadline is met.
	CallUnicast(
		ctx context.Context,
		msgID overpass.MessageID,
		target overpass.PeerID,
		namespace string,
		command string,
		payload *overpass.Payload,
	) (*overpass.Payload, error)

	// CallBalanced sends a load-balanced command request to the first available
	// peer and blocks until a response is received or the context deadline is met.
	CallBalanced(
		ctx context.Context,
		msgID overpass.MessageID,
		namespace string,
		command string,
		payload *overpass.Payload,
	) (*overpass.Payload, error)

	// ExecuteBalanced sends a load-balanced command request to the first
	// available peer and returns immediately.
	ExecuteBalanced(
		ctx context.Context,
		msgID overpass.MessageID,
		namespace string,
		command string,
		payload *overpass.Payload,
	) error

	// ExecuteMulticast sends a multicast command request to the all available
	// peers and returns immediately.
	ExecuteMulticast(
		ctx context.Context,
		msgID overpass.MessageID,
		namespace string,
		command string,
		payload *overpass.Payload,
	) error
}
