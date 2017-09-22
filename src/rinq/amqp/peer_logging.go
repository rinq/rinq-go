package amqp

import (
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

func logStartedListening(
	logger rinq.Logger,
	peerID ident.PeerID,
	namespace string,
) {
	logger.Log(
		"%s started listening for command requests in '%s' namespace",
		peerID.ShortString(),
		namespace,
	)
}

func logStoppedListening(
	logger rinq.Logger,
	peerID ident.PeerID,
	namespace string,
) {
	logger.Log(
		"%s stopped listening for command requests in '%s' namespace",
		peerID.ShortString(),
		namespace,
	)
}
