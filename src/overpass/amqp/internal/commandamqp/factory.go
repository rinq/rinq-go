package commandamqp

import (
	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpass/amqp/internal/amqputil"
	"github.com/over-pass/overpass-go/src/overpass/internal/command"
	"github.com/over-pass/overpass-go/src/overpass/internal/revision"
)

// New returns a pair of invoker and server.
func New(
	peerID overpass.PeerID,
	config overpass.Config,
	revisions revision.Store,
	channels amqputil.ChannelPool,
) (command.Invoker, command.Server, error) {
	channel, err := channels.Get()
	if err != nil {
		return nil, nil, err
	}
	defer channels.Put(channel)

	if err = declareExchanges(channel); err != nil {
		return nil, nil, err
	}

	queues := &queueSet{}

	invoker, err := newInvoker(
		peerID,
		config.SessionPreFetch,
		config.DefaultTimeout,
		queues,
		channels,
		config.Logger,
	)
	if err != nil {
		return nil, nil, err
	}

	server, err := newServer(
		peerID,
		config.CommandPreFetch,
		revisions,
		queues,
		channels,
		config.Logger,
	)
	if err != nil {
		invoker.Stop()
		<-invoker.Done()
		return nil, nil, err
	}

	return invoker, server, nil
}
