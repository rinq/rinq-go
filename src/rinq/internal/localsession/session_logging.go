package localsession

import (
	"context"
	"time"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
	"github.com/rinq/rinq-go/src/rinq/internal/bufferpool"
	"github.com/rinq/rinq-go/src/rinq/internal/trace"
)

func logCall(
	logger rinq.Logger,
	msgID ident.MessageID,
	ns string,
	cmd string,
	elapsed time.Duration,
	out *rinq.Payload,
	in *rinq.Payload,
	err error,
	traceID string,
) {
	switch e := err.(type) {
	case nil:
		logger.Log(
			"%s called '%s::%s' command: success (%dms, %d/o %d/i) [%s]",
			msgID.ShortString(),
			ns,
			cmd,
			elapsed,
			out.Len(),
			in.Len(),
			traceID,
		)
	case rinq.Failure:
		logger.Log(
			"%s called '%s::%s' command: '%s' failure (%dms, %d/o %d/i) [%s]",
			msgID.ShortString(),
			ns,
			cmd,
			e.Type,
			elapsed,
			out.Len(),
			in.Len(),
			traceID,
		)
	case rinq.CommandError:
		logger.Log(
			"%s called '%s::%s' command: '%s' error (%dms, %d/o 0/i) [%s]",
			msgID.ShortString(),
			ns,
			cmd,
			e,
			elapsed,
			out.Len(),
			traceID,
		)
	default:
		if err == context.DeadlineExceeded || err == context.Canceled {
			logger.Log(
				"%s called '%s::%s' command: %s (%dms, %d/o -/i) [%s]",
				msgID.ShortString(),
				ns,
				cmd,
				err,
				elapsed,
				out.Len(),
				traceID,
			)
		}
	}
}

func logAsyncRequest(
	logger rinq.Logger,
	msgID ident.MessageID,
	ns string,
	cmd string,
	out *rinq.Payload,
	traceID string,
) {
	logger.Log(
		"%s called '%s::%s' command asynchronously (%d/o) [%s]",
		msgID.ShortString(),
		ns,
		cmd,
		out.Len(),
		traceID,
	)
}

func logAsyncResponse(
	ctx context.Context,
	logger rinq.Logger,
	msgID ident.MessageID,
	ns string,
	cmd string,
	in *rinq.Payload,
	err error,
) {
	switch e := err.(type) {
	case nil:
		logger.Log(
			"%s called '%s::%s' command asynchronously: success (%d/i) [%s]",
			msgID.ShortString(),
			ns,
			cmd,
			in.Len(),
			trace.Get(ctx),
		)
	case rinq.Failure:
		logger.Log(
			"%s called '%s::%s' command asynchronously: '%s' failure (%d/i) [%s]",
			msgID.ShortString(),
			ns,
			cmd,
			e.Type,
			in.Len(),
			trace.Get(ctx),
		)
	case rinq.CommandError:
		logger.Log(
			"%s called '%s::%s' command asynchronously: '%s' error (0/i) [%s]",
			msgID.ShortString(),
			ns,
			cmd,
			e,
			trace.Get(ctx),
		)
	}
}

func logSessionClose(
	logger rinq.Logger,
	cat Catalog,
	traceID string,
) {
	ref, attrs := cat.Attrs()

	buffer := bufferpool.Get()
	defer bufferpool.Put(buffer)
	attrmeta.WriteTable(buffer, attrs)

	if traceID == "" {
		logger.Log(
			"%s session destroyed {%s}",
			ref.ShortString(),
			buffer,
		)
	} else {
		logger.Log(
			"%s session destroyed {%s} [%s]",
			ref.ShortString(),
			buffer,
			traceID,
		)
	}
}
