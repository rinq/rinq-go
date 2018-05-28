package remotesession

import (
	"context"

	"github.com/jmalloc/twelf/src/twelf"
	"github.com/rinq/rinq-go/src/internal/attributes"
	"github.com/rinq/rinq-go/src/internal/localsession"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/trace"
)

func logRemoteUpdate(
	ctx context.Context,
	logger twelf.Logger,
	ref ident.Ref,
	peerID ident.PeerID,
	diff *attributes.Diff,
) {
	logger.Log(
		"%s session updated by %s %s [%s]",
		ref.ShortString(),
		peerID.ShortString(),
		diff,
		trace.Get(ctx),
	)
}

func logRemoteClear(
	ctx context.Context,
	logger twelf.Logger,
	ref ident.Ref,
	peerID ident.PeerID,
	diff *attributes.Diff,
) {
	logger.Log(
		"%s session cleared by %s %s [%s]",
		ref.ShortString(),
		peerID.ShortString(),
		diff,
		trace.Get(ctx),
	)
}

func logRemoteDestroy(
	ctx context.Context,
	logger twelf.Logger,
	sess *localsession.Session,
	peerID ident.PeerID,
) {
	ref, attrs := sess.Attrs()

	logger.Log(
		"%s session destroyed by %s %s [%s]",
		ref.ShortString(),
		peerID.ShortString(),
		attrs,
		trace.Get(ctx),
	)
}
