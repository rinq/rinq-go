package notifyamqp

import "github.com/over-pass/overpass-go/src/overpass"

func logListenerStart(
	logger overpass.Logger,
	peerID overpass.PeerID,
	preFetch int,
) {
	if !logger.IsDebug() {
		return
	}

	logger.Log(
		"%s notification listener started with pre-fetch of %d message(s)",
		peerID.ShortString(),
		preFetch,
	)
}

func logListenerStopping(
	logger overpass.Logger,
	peerID overpass.PeerID,
) {
	if !logger.IsDebug() {
		return
	}

	logger.Log(
		"%s notification listener is stopping gracefully",
		peerID.ShortString(),
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
			"%s notification listener stopped",
			peerID.ShortString(),
		)
	} else {
		logger.Log(
			"%s notification listener stopped with error: %s",
			peerID.ShortString(),
			err,
		)
	}
}
