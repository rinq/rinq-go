package remotesession

import (
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

func logCacheAdd(
	logger rinq.Logger,
	peerID ident.PeerID,
	sessID ident.SessionID,
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
	peerID ident.PeerID,
	sessID ident.SessionID,
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
	peerID ident.PeerID,
	sessID ident.SessionID,
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
