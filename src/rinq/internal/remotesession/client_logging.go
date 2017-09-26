package remotesession

import (
	"bytes"
	"context"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/trace"
)

func logUpdate(
	ctx context.Context,
	logger rinq.Logger,
	peerID ident.PeerID,
	ref ident.Ref,
	diff *bytes.Buffer,
) {
	logger.Log(
		"%s updated remote session %s {%s} [%s]",
		peerID.ShortString(),
		ref.ShortString(),
		diff.String(),
		trace.Get(ctx),
	)
}

func logClose(
	ctx context.Context,
	logger rinq.Logger,
	peerID ident.PeerID,
	ref ident.Ref,
) {
	logger.Log(
		"%s destroyed remote session %s [%s]",
		peerID.ShortString(),
		ref.ShortString(),
		trace.Get(ctx),
	)
}
