package notifications

import (
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/streadway/amqp"
)

func logConsumerStart(
	l rinq.Logger,
	p ident.PeerID,
	preFetch int,
) {
	if l.IsDebug() {
		l.Log(
			"%s notification consumer started (pre-fetch: %d)",
			p.ShortString(),
			preFetch,
		)
	}
}

func logConsumerStop(
	l rinq.Logger,
	p ident.PeerID,
	err error,
) {
	if l.IsDebug() {
		if err == nil {
			l.Log(
				"%s notification consumer stopped",
				p.ShortString(),
			)
		} else {
			l.Log(
				"%s notification consumer stopped: %s",
				p.ShortString(),
				err,
			)
		}
	}
}

func logIgnoredMessage(
	logger rinq.Logger,
	peerID ident.PeerID,
	msg *amqp.Delivery,
	err error,
) {
	if logger.IsDebug() {
		logger.Log(
			"%s ignored AMQP notification message %s, %s",
			peerID.ShortString(),
			msg.MessageId,
			err,
		)
	}
}
