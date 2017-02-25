package notifyamqp

import "github.com/over-pass/overpass-go/src/overpass"

func logInvalidMessageID(
	logger overpass.Logger,
	peerID overpass.PeerID,
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
	logger overpass.Logger,
	peerID overpass.PeerID,
	msgID overpass.MessageID,
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
	logger overpass.Logger,
	peerID overpass.PeerID,
	preFetch int,
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
	logger overpass.Logger,
	peerID overpass.PeerID,
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
	logger overpass.Logger,
	peerID overpass.PeerID,
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
