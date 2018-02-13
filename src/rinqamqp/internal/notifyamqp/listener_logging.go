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
	if !logger.IsDebug() {
		return
	}

	logger.Log(
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
	if !logger.IsDebug() {
		return
	}

	logger.Log(
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
	if !logger.IsDebug() {
		return
	}

	logger.Log(
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
	if !logger.IsDebug() {
		return
	}

	logger.Log(
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
	if !logger.IsDebug() {
		return
	}

	if err == nil {
		logger.Log(
			"%s listener stopped",
			peerID.ShortString(),
		)
	} else {
		logger.Log(
			"%s listener stopped: %s",
			peerID.ShortString(),
			err,
		)
	}
}
