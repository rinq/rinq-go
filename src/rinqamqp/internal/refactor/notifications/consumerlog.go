package notifications

import (
	"github.com/jmalloc/twelf/src/twelf"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/streadway/amqp"
)

func logConsumerStart(
	l twelf.Logger,
	p ident.PeerID,
	preFetch int,
) {
	l.Debug(
		"%s notification consumer started (pre-fetch: %d)",
		p.ShortString(),
		preFetch,
	)
}

func logConsumerStop(
	l twelf.Logger,
	p ident.PeerID,
	err error,
) {
	if err == nil {
		l.Debug(
			"%s notification consumer stopped",
			p.ShortString(),
		)
	} else {
		l.Debug(
			"%s notification consumer stopped: %s",
			p.ShortString(),
			err,
		)
	}
}

func logIgnoredMessage(
	logger twelf.Logger,
	peerID ident.PeerID,
	msg *amqp.Delivery,
	err error,
) {
	logger.Debug(
		"%s ignored AMQP notification message %s, %s",
		peerID.ShortString(),
		msg.MessageId,
		err,
	)
}
