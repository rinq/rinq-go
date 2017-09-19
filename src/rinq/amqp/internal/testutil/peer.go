package testutil

import (
	"sync"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/amqp"
	"github.com/rinq/rinq-go/src/rinq/options"
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
	peer, err := amqp.DialEnv(
		options.Logger(rinq.NewLogger(true)),
	)

	if err != nil {
		panic(err)
	}

	return peer
}
