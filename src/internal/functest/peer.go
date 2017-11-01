package functest

import (
	"sync"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/options"
	"github.com/rinq/rinq-go/src/rinqamqp"
)

var peers struct {
	mutex  sync.Mutex
	shared rinq.Peer
	all    []rinq.Peer
}

// SharedPeer returns a peer for use in functional tests.
func SharedPeer() rinq.Peer {
	peer := NewPeer()

	peers.mutex.Lock()
	defer peers.mutex.Unlock()

	if peers.shared == nil {
		peers.shared = peer
	}

	return peers.shared
}

// NewPeer returns a new peer for use in functional tests.
func NewPeer() rinq.Peer {
	peer, err := rinqamqp.DialEnv(
		options.Logger(rinq.NewLogger(true)),
	)

	if err != nil {
		panic(err)
	}

	peers.mutex.Lock()
	defer peers.mutex.Unlock()

	peers.all = append(peers.all, peer)

	return peer
}

func tearDownPeers() {
	peers.mutex.Lock()
	defer peers.mutex.Unlock()

	peers.shared = nil

	var wg sync.WaitGroup

	for _, peer := range peers.all {
		wg.Add(1)

		go func(p rinq.Peer) {
			p.Stop()
			<-p.Done()
			wg.Done()
		}(peer)
	}

	wg.Wait()
	peers.all = nil
}
