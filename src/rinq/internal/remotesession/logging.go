package remotesession

import "github.com/rinq/rinq-go/src/rinq"

func logCacheAdd(
	logger rinq.Logger,
	peerID rinq.PeerID,
	sessID rinq.SessionID,
) {
	if !logger.IsDebug() {
		return
	}

	logger.Log(
		"%s discovered remote session %s ",
		peerID.ShortString(),
		sessID.ShortString(),
	)
}

func logCacheMark(
	logger rinq.Logger,
	peerID rinq.PeerID,
	sessID rinq.SessionID,
) {
	if !logger.IsDebug() {
		return
	}

	logger.Log(
		"%s marked remote session %s for removal from the store",
		peerID.ShortString(),
		sessID.ShortString(),
	)
}

func logCacheRemove(
	logger rinq.Logger,
	peerID rinq.PeerID,
	sessID rinq.SessionID,
) {
	if !logger.IsDebug() {
		return
	}

	logger.Log(
		"%s removed remote session %s from the store",
		peerID.ShortString(),
		sessID.ShortString(),
	)
}
