package notifyamqp

import (
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/amqp/internal/amqputil"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/localsession"
	"github.com/rinq/rinq-go/src/rinq/internal/notify"
	"github.com/rinq/rinq-go/src/rinq/internal/revision"
)

// New returns a pair of notifier and listener.
func New(
	peerID ident.PeerID,
	config rinq.Config,
	sessions localsession.Store,
	revisions revision.Store,
	channels amqputil.ChannelPool,
) (notify.Notifier, notify.Listener, error) {
	channel, err := channels.GetQOS(config.SessionPreFetch) // do not return to pool, use for listener
	if err != nil {
		return nil, nil, err
	}

	if err = declareExchanges(channel); err != nil {
		return nil, nil, err
	}

	listener, err := newListener(
		peerID,
		config.SessionPreFetch,
		sessions,
		revisions,
		channel,
		config.Logger,
	)
	if err != nil {
		return nil, nil, err
	}

	return newNotifier(channels, config.Logger), listener, nil
}