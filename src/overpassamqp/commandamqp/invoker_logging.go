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
		"%s invoker started (pre-fetch: %d)",
		peerID.ShortString(),
		preFetch,
	)
}

func logInvokerStopping(
	logger overpass.Logger,
	peerID overpass.PeerID,
	pending int,
) {
	if !logger.IsDebug() {
		return
	}

	logger.Log(
		"%s invoker stopping gracefully (pending: %d)",
		peerID.ShortString(),
		pending,
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
			"%s invoker stopped",
			peerID.ShortString(),
		)
	} else {
		logger.Log(
			"%s invoker stopped: %s",
			peerID.ShortString(),
			err,
		)
	}
}
