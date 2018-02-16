package commandamqp

import (
	"github.com/rinq/rinq-go/src/internal/command"
	"github.com/rinq/rinq-go/src/internal/localsession"
	"github.com/rinq/rinq-go/src/internal/revisions"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/options"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/amqputil"
)

// New returns a pair of invoker and server.
func New(
	peerID ident.PeerID,
	opts options.Options,
	sessions *localsession.Store,
	revs revisions.Store,
	channels amqputil.ChannelPool,
	queues *QueueSet,
) (command.Invoker, command.Server, error) {
	channel, err := channels.Get()
	if err != nil {
		return nil, nil, err
	}
	defer channels.Put(channel)

	if err = declareExchanges(channel); err != nil {
		return nil, nil, err
	}

	invoker, err := newInvoker(
		peerID,
		opts.SessionWorkers,
		opts.DefaultTimeout,
		sessions,
		queues,
		channels,
		opts.Logger,
		opts.Tracer,
	)
	if err != nil {
		return nil, nil, err
	}

	server, err := newServer(
		peerID,
		opts.CommandWorkers,
		revs,
		queues,
		channels,
		opts.Logger,
		opts.Tracer,
	)
	if err != nil {
		invoker.Stop()
		<-invoker.Done()
		return nil, nil, err
	}

	return invoker, server, nil
}
