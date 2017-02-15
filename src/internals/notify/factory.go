package notify

import (
	"github.com/over-pass/overpass-go/src/internals"
	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/over-pass/overpass-go/src/overpass"
)

// New returns a pair of notifier and listener.
func New(
	peerID overpass.PeerID,
	config overpass.Config,
	sessions internals.SessionStore,
	revisions internals.RevisionStore,
	channels amqputil.ChannelPool,
) (internals.Notifier, internals.Listener, error) {
	channel, err := channels.Get() // do not return to pool, use for listener
	if err != nil {
		return nil, nil, err
	}

	if err = declareExchanges(channel); err != nil {
		return nil, nil, err
	}

	listener, err := newListener(
		peerID,
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
