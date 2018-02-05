package notifications

import (
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/streadway/amqp"
)

func logSourceStart(
	logger rinq.Logger,
	peerID ident.PeerID,
	preFetch int,
) {
	if logger.IsDebug() {
		logger.Log(
			"%s notification source started (pre-fetch: %d)",
			peerID.ShortString(),
			preFetch,
		)
	}
}

func logSourceStop(
	logger rinq.Logger,
	peerID ident.PeerID,
	err error,
) {
	if logger.IsDebug() {
		if err == nil {
			logger.Log(
				"%s notification source stopped",
				peerID.ShortString(),
			)
		} else {
			logger.Log(
				"%s notification source stopped: %s",
				peerID.ShortString(),
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
			"%s notification source ignored AMQP message %s, %s",
			peerID.ShortString(),
			msg.MessageId,
			err,
		)
	}
}

func logIgnoredSpanContext(
	logger rinq.Logger,
	peerID ident.PeerID,
	msg *amqp.Delivery,
	err error,
) {
	if logger.IsDebug() {
		logger.Log(
			"%s notification source ignored invalid span context in AMQP message %s, %s",
			peerID.ShortString(),
			msg.MessageId,
			err,
		)
	}
}
