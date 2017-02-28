package notifyamqp

import (
	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpass/amqp/internal/amqputil"
	"github.com/over-pass/overpass-go/src/overpass/internal/localsession"
	"github.com/over-pass/overpass-go/src/overpass/internal/notify"
	"github.com/over-pass/overpass-go/src/overpass/internal/revision"
)

// New returns a pair of notifier and listener.
func New(
	peerID overpass.PeerID,
	config overpass.Config,
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
