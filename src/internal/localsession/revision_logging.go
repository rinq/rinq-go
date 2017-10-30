package localsession

import (
	"context"

	"github.com/rinq/rinq-go/src/internal/attributes"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/trace"
)

func logUpdate(
	ctx context.Context,
	logger rinq.Logger,
	ref ident.Ref,
	diff *attributes.Diff,
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

func logClear(
	ctx context.Context,
	logger rinq.Logger,
	ref ident.Ref,
	diff *attributes.Diff,
) {
	if traceID := trace.Get(ctx); traceID != "" {
		logger.Log(
			"%s session cleared %s [%s]",
			ref.ShortString(),
			diff,
			traceID,
		)
	} else {
		logger.Log(
			"%s session cleared %s",
			ref.ShortString(),
			diff,
		)
	}
}

func logDestroy(
	ctx context.Context,
	logger rinq.Logger,
	state *state,
) {
	logSessionDestroy(logger, state, trace.Get(ctx))
}
