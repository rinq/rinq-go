package commandamqp

import "github.com/over-pass/overpass-go/src/overpass"

func logInvokerStart(
	logger overpass.Logger,
	peerID overpass.PeerID,
	preFetch int,
) {
	if !logger.IsDebug() {
		return
	}

	logger.Log(
		"%s command invoker started with pre-fetch of %d message(s)",
		peerID.ShortString(),
		preFetch,
	)
}

func logInvokerStopping(
	logger overpass.Logger,
	peerID overpass.PeerID,
) {
	if !logger.IsDebug() {
		return
	}

	logger.Log(
		"%s command invoker is stopping gracefully",
		peerID.ShortString(),
	)
}

func logInvokerStop(
	logger overpass.Logger,
	peerID overpass.PeerID,
	err error,
) {
	if !logger.IsDebug() {
		return
	}

	if err == nil {
		logger.Log(
			"%s command invoker stopped",
			peerID.ShortString(),
		)
	} else {
		logger.Log(
			"%s command invoker stopped with error: %s",
			peerID.ShortString(),
			err,
		)
	}
}
