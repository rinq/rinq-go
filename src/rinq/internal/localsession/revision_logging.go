package localsession

import (
	"context"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
	"github.com/rinq/rinq-go/src/rinq/trace"
)

func logUpdate(
	ctx context.Context,
	logger rinq.Logger,
	ref ident.Ref,
	diff *attrmeta.Diff,
) {
	if traceID := trace.Get(ctx); traceID != "" {
		logger.Log(
			"%s session updated %s [%s]",
			ref.ShortString(),
			diff,
			traceID,
		)
	} else {
		logger.Log(
			"%s session updated %s",
			ref.ShortString(),
			diff,
		)
	}
}

func logDestroy(
	ctx context.Context,
	logger rinq.Logger,
	cat Catalog,
) {
	logSessionDestroy(logger, cat, trace.Get(ctx))
}
