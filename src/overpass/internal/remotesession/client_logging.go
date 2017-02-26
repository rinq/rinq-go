package remotesession

import (
	"context"

	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpass/internal/attrmeta"
	"github.com/over-pass/overpass-go/src/overpass/internal/bufferpool"
	"github.com/over-pass/overpass-go/src/overpass/internal/trace"
)

func logUpdate(
	ctx context.Context,
	logger overpass.Logger,
	peerID overpass.PeerID,
	ref overpass.SessionRef,
	attrs []attrmeta.Attr,
) {
	buffer := bufferpool.Get()
	defer bufferpool.Put(buffer)

	for _, attr := range attrs {
		if buffer.Len() != 0 {
			buffer.WriteString(", ")
		}

		attrmeta.Write(buffer, attr)
	}

	logger.Log(
		"%s updated remote session %s {%s} [%s]",
		peerID.ShortString(),
		ref.ShortString(),
		buffer.String(),
		trace.Get(ctx),
	)
}

func logClose(
	ctx context.Context,
	logger overpass.Logger,
	peerID overpass.PeerID,
	ref overpass.SessionRef,
) {
	logger.Log(
		"%s destroyed remote session %s [%s]",
		peerID.ShortString(),
		ref.ShortString(),
		trace.Get(ctx),
	)
}
