package remotesession

import (
	"context"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
	"github.com/rinq/rinq-go/src/rinq/internal/localsession"
	"github.com/rinq/rinq-go/src/rinq/trace"
)

func logRemoteUpdate(
	ctx context.Context,
	logger rinq.Logger,
	ref ident.Ref,
	peerID ident.PeerID,
	diff *attrmeta.Diff,
) {
	logger.Log(
		"%s session updated by %s %s [%s]",
		ref.ShortString(),
		peerID.ShortString(),
		diff,
		trace.Get(ctx),
	)
}

func logRemoteDestroy(
	ctx context.Context,
	logger rinq.Logger,
	cat localsession.Catalog,
	peerID ident.PeerID,
) {
	ref, attrs := cat.Attrs()

	logger.Log(
		"%s session destroyed by %s %s [%s]",
		ref.ShortString(),
		peerID.ShortString(),
		attrs,
		trace.Get(ctx),
	)
}
