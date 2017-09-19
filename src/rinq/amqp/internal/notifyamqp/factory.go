package notifyamqp

import (
	"github.com/rinq/rinq-go/src/rinq/amqp/internal/amqputil"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/localsession"
	"github.com/rinq/rinq-go/src/rinq/internal/notify"
	"github.com/rinq/rinq-go/src/rinq/internal/optutil"
	"github.com/rinq/rinq-go/src/rinq/internal/revision"
)

// New returns a pair of notifier and listener.
func New(
	peerID ident.PeerID,
	cfg optutil.Config,
	sessions localsession.Store,
	revisions revision.Store,
	channels amqputil.ChannelPool,
) (notify.Notifier, notify.Listener, error) {
	channel, err := channels.GetQOS(cfg.SessionWorkers) // do not return to pool, use for listener
	if err != nil {
		return nil, nil, err
	}

	if err = declareExchanges(channel); err != nil {
		return nil, nil, err
	}

	listener, err := newListener(
		peerID,
		cfg.SessionWorkers,
		sessions,
		revisions,
		channel,
		cfg.Logger,
	)
	if err != nil {
		return nil, nil, err
	}

	return newNotifier(peerID, channels, cfg.Logger), listener, nil
}
