package notifyamqp

import (
	"github.com/jmalloc/twelf/src/twelf"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

func logNotifierStart(
	logger twelf.Logger,
	peerID ident.PeerID,
) {
	if !logger.IsDebug() {
		return
	}

	logger.Log(
		"%s notifier started",
		peerID.ShortString(),
	)
}

func logNotifierStop(
	logger twelf.Logger,
	peerID ident.PeerID,
	err error,
) {
	if !logger.IsDebug() {
		return
	}

	if err == nil {
		logger.Log(
			"%s notifier stopped",
			peerID.ShortString(),
		)
	} else {
		logger.Log(
			"%s notifier stopped: %s",
			peerID.ShortString(),
			err,
		)
	}
}
