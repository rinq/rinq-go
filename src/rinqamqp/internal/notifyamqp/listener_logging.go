package notifyamqp

import (
	"github.com/jmalloc/twelf/src/twelf"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

func logInvalidMessageID(
	logger twelf.Logger,
	peerID ident.PeerID,
	msgID string,
) {
	logger.Debug(
		"%s listener ignored AMQP message, '%s' is not a valid message ID",
		peerID.ShortString(),
		msgID,
	)
}

func logIgnoredMessage(
	logger twelf.Logger,
	peerID ident.PeerID,
	msgID ident.MessageID,
	err error,
) {
	logger.Debug(
		"%s listener ignored AMQP message %s, %s",
		peerID.ShortString(),
		msgID.ShortString(),
		err,
	)
}

func logListenerStart(
	logger twelf.Logger,
	peerID ident.PeerID,
	preFetch uint,
) {
	logger.Debug(
		"%s listener started (pre-fetch: %d)",
		peerID.ShortString(),
		preFetch,
	)
}

func logListenerStopping(
	logger twelf.Logger,
	peerID ident.PeerID,
	pending uint,
) {
	logger.Debug(
		"%s listener stopping gracefully (pending: %d)",
		peerID.ShortString(),
		pending,
	)
}

func logListenerStop(
	logger twelf.Logger,
	peerID ident.PeerID,
	err error,
) {
	if err == nil {
		logger.Debug(
			"%s listener stopped",
			peerID.ShortString(),
		)
	} else {
		logger.Debug(
			"%s listener stopped: %s",
			peerID.ShortString(),
			err,
		)
	}
}
