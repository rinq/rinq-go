package remotesession

import (
	"context"

	"github.com/jmalloc/twelf/src/twelf"
	"github.com/rinq/rinq-go/src/internal/attributes"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/trace"
)

func logUpdate(
	ctx context.Context,
	logger twelf.Logger,
	peerID ident.PeerID,
	ref ident.Ref,
	diff *attributes.Diff,
) {
	logger.Log(
		"%s updated remote session %s %s [%s]",
		peerID.ShortString(),
		ref.ShortString(),
		diff,
		trace.Get(ctx),
	)
}

func logClear(
	ctx context.Context,
	logger twelf.Logger,
	peerID ident.PeerID,
	ref ident.Ref,
	ns string,
) {
	logger.Log(
		"%s cleared remote session %s %s::{*} [%s]",
		peerID.ShortString(),
		ref.ShortString(),
		trace.Get(ctx),
		ns,
	)
}

func logClose(
	ctx context.Context,
	logger twelf.Logger,
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
