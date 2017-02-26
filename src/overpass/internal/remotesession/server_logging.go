package remotesession

import (
	"bytes"
	"context"

	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpass/internal/attrmeta"
	"github.com/over-pass/overpass-go/src/overpass/internal/bufferpool"
	"github.com/over-pass/overpass-go/src/overpass/internal/localsession"
	"github.com/over-pass/overpass-go/src/overpass/internal/trace"
)

func logRemoteUpdate(
	ctx context.Context,
	logger overpass.Logger,
	ref overpass.SessionRef,
	peerID overpass.PeerID,
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
	logger overpass.Logger,
	cat localsession.Catalog,
	peerID overpass.PeerID,
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
