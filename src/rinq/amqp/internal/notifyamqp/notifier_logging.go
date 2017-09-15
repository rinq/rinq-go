package notifyamqp

import (
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

func logNotifierStart(
	logger rinq.Logger,
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
	logger rinq.Logger,
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
