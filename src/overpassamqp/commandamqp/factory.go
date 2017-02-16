package commandamqp

import (
	"github.com/over-pass/overpass-go/src/internals"
	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/over-pass/overpass-go/src/internals/command"
	"github.com/over-pass/overpass-go/src/overpass"
)

// New returns a pair of invoker and server.
func New(
	peerID overpass.PeerID,
	config overpass.Config,
	revisions internals.RevisionStore,
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
		config.PreFetch,
		revisions,
		queues,
		channels,
		config.Logger,
	)
	if err != nil {
		// TODO: invoker.Stop()
		return nil, nil, err
	}

	return invoker, server, nil
}
