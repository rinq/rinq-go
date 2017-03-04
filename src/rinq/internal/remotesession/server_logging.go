package remotesession

import (
	"bytes"
	"context"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
	"github.com/rinq/rinq-go/src/rinq/internal/bufferpool"
	"github.com/rinq/rinq-go/src/rinq/internal/localsession"
	"github.com/rinq/rinq-go/src/rinq/internal/trace"
)

func logRemoteUpdate(
	ctx context.Context,
	logger rinq.Logger,
	ref ident.Ref,
	peerID ident.PeerID,
	diff *bytes.Buffer,
) {
	logger.Log(
		"%s session updated by %s {%s} [%s]",
		ref.ShortString(),
		peerID.ShortString(),
		diff.String(),
		trace.Get(ctx),
	)
}

func logRemoteClose(
	ctx context.Context,
	logger rinq.Logger,
	cat localsession.Catalog,
	peerID ident.PeerID,
) {
	ref, attrs := cat.Attrs()

	buffer := bufferpool.Get()
	defer bufferpool.Put(buffer)
	attrmeta.WriteTable(buffer, attrs)

	logger.Log(
		"%s session destroyed by %s {%s} [%s]",
		ref.ShortString(),
		peerID.ShortString(),
		buffer,
		trace.Get(ctx),
	)
}
