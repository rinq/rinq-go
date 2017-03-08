package commandamqp

import (
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/amqp/internal/amqputil"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/command"
	"github.com/rinq/rinq-go/src/rinq/internal/localsession"
	"github.com/rinq/rinq-go/src/rinq/internal/revision"
)

// New returns a pair of invoker and server.
func New(
	peerID ident.PeerID,
	config rinq.Config,
	sessions localsession.Store,
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
		sessions,
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
