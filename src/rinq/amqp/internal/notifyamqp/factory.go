package notifyamqp

import (
	"github.com/rinq/rinq-go/src/internal/localsession"
	"github.com/rinq/rinq-go/src/internal/notify"
	"github.com/rinq/rinq-go/src/internal/revision"
	"github.com/rinq/rinq-go/src/rinq/amqp/internal/amqputil"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/options"
)

// New returns a pair of notifier and listener.
func New(
	peerID ident.PeerID,
	opts options.Options,
	sessions localsession.Store,
	revisions revision.Store,
	channels amqputil.ChannelPool,
) (notify.Notifier, notify.Listener, error) {
	channel, err := channels.GetQOS(opts.SessionWorkers) // do not return to pool, use for listener
	if err != nil {
		return nil, nil, err
	}

	if err = declareExchanges(channel); err != nil {
		return nil, nil, err
	}

	listener, err := newListener(
		peerID,
		opts.SessionWorkers,
		sessions,
		revisions,
		channel,
		opts.Logger,
		opts.Tracer,
	)
	if err != nil {
		return nil, nil, err
	}

	return newNotifier(peerID, channels, opts.Logger), listener, nil
}
