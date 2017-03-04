package localsession

import (
	"bytes"
	"context"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/trace"
)

func logUpdate(
	ctx context.Context,
	logger rinq.Logger,
	ref ident.Ref,
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

func logClose(
	ctx context.Context,
	logger rinq.Logger,
	cat Catalog,
) {
	logSessionClose(logger, cat, trace.Get(ctx))
}
