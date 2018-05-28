package rinqamqp

import (
	"github.com/jmalloc/twelf/src/twelf"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

func logStartedListening(
	logger twelf.Logger,
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
	logger twelf.Logger,
	peerID ident.PeerID,
	namespace string,
) {
	logger.Log(
		"%s stopped listening for command requests in '%s' namespace",
		peerID.ShortString(),
		namespace,
	)
}
