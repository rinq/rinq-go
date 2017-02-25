package commandamqp

import (
	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/over-pass/overpass-go/src/internals/command"
	"github.com/over-pass/overpass-go/src/internals/revision"
	"github.com/over-pass/overpass-go/src/overpass"
)

// New returns a pair of invoker and server.
func New(
	peerID overpass.PeerID,
	config overpass.Config,
	revisions revision.Store,
	channels amqputil.ChannelPool,
) (command.Invoker, command.Server, error) {
	channel, err := channels.Get() // do not return to pool, used for invoker
	if err != nil {
		return nil, nil, err
	}

	if err = declareExchanges(channel); err != nil {
		return nil, nil, err
	}

	queues := &queueSet{}

	invoker, err := newInvoker(
		peerID,
		config.SessionPreFetch,
		config.DefaultTimeout,
		queues,
		channel,
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
		return nil, nil, err
	}

	return invoker, server, nil
}
