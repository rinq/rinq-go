package functest

import (
	"sync"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/options"
	"github.com/rinq/rinq-go/src/rinqamqp"
)

var sharedPeer struct {
	mutex sync.Mutex
	peer  rinq.Peer
}

// SharedPeer returns a peer for use in functional tests.
func SharedPeer() rinq.Peer {
	sharedPeer.mutex.Lock()
	defer sharedPeer.mutex.Unlock()

	if sharedPeer.peer == nil {
		sharedPeer.peer = NewPeer()
	}

	return sharedPeer.peer
}

// NewPeer returns a new peer for use in functional tests.
func NewPeer() rinq.Peer {
	peer, err := rinqamqp.DialEnv(
		options.Logger(rinq.NewLogger(true)),
	)

	if err != nil {
		panic(err)
	}

	return peer
}
