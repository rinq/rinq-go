package localsession

import (
	"bytes"
	"context"

	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpass/internal/trace"
)

func logUpdate(
	ctx context.Context,
	logger overpass.Logger,
	ref overpass.SessionRef,
	diff *bytes.Buffer,
) {
	if traceID := trace.Get(ctx); traceID != "" {
		logger.Log(
			"%s session updated {%s} [%s]",
			ref.ShortString(),
			diff.String(),
			traceID,
		)
	} else {
		logger.Log(
			"%s session updated {%s}",
			ref.ShortString(),
			diff.String(),
		)
	}
}
