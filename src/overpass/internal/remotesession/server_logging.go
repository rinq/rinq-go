package remotesession

import (
	"bytes"
	"context"

	"github.com/over-pass/overpass-go/src/overpass"
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
