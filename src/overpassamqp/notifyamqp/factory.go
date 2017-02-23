package notifyamqp

import (
	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/over-pass/overpass-go/src/internals/localsession"
	"github.com/over-pass/overpass-go/src/internals/notify"
	"github.com/over-pass/overpass-go/src/internals/revision"
	"github.com/over-pass/overpass-go/src/overpass"
)

// New returns a pair of notifier and listener.
func New(
	peerID overpass.PeerID,
	config overpass.Config,
	sessions localsession.Store,
	revisions revision.Store,
	channels amqputil.ChannelPool,
) (notify.Notifier, notify.Listener, error) {
	channel, err := channels.Get() // do not return to pool, use for listener
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
