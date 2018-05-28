package notifyamqp

import (
	"github.com/jmalloc/twelf/src/twelf"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

func logNotifierStart(
	logger twelf.Logger,
	peerID ident.PeerID,
) {
	logger.Debug(
		"%s notifier started",
		peerID.ShortString(),
	)
}

func logNotifierStop(
	logger twelf.Logger,
	peerID ident.PeerID,
	err error,
) {
	if err == nil {
		logger.Debug(
			"%s notifier stopped",
			peerID.ShortString(),
		)
	} else {
		logger.Debug(
			"%s notifier stopped: %s",
			peerID.ShortString(),
			err,
		)
	}
}
