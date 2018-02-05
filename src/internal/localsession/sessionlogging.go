package localsession

import (
	"context"
	"time"

	"github.com/jmalloc/twelf/src/twelf"
	"github.com/rinq/rinq-go/src/internal/attributes"
	"github.com/rinq/rinq-go/src/internal/notifications"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/trace"
)

func logCreated(
	logger twelf.Logger,
	id ident.SessionID,
) {
	logger.Log(
		"%s session created",
		id.At(0).ShortString(),
	)
}

func logCall(
	logger twelf.Logger,
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
	logger twelf.Logger,
	msgID ident.MessageID,
	ns string,
	cmd string,
	out *rinq.Payload,
	err error,
	traceID string,
) {
	if err != nil {
		return // request never sent
	}

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
	logger twelf.Logger,
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

func logExecute(
	logger twelf.Logger,
	msgID ident.MessageID,
	ns string,
	cmd string,
	out *rinq.Payload,
	err error,
	traceID string,
) {
	if err != nil {
		return // request never sent
	}

	logger.Log(
		"%s executed '%s::%s' command (%d/o) [%s]",
		msgID.ShortString(),
		ns,
		cmd,
		out.Len(),
		traceID,
	)
}

func logNotify(logger twelf.Logger, n *notifications.Notification) {
	logger.Log(
		"%s sent '%s::%s' notification to %s (%d/o) [%s]",
		n.ID.ShortString(),
		n.Namespace,
		n.Type,
		n.UnicastTarget.ShortString(),
		n.Payload.Len(),
		n.TraceID,
	)
}

func logNotifyMany(logger twelf.Logger, n *notifications.Notification) {
	logger.Log(
		"%s sent '%s::%s' notification to sessions matching %s (%d/o) [%s]",
		n.ID.ShortString(),
		n.Namespace,
		n.Type,
		n.MulticastConstraint,
		n.Payload.Len(),
		n.TraceID,
	)
}

func logNotifyRecv(
	logger twelf.Logger,
	ref ident.Ref,
	n rinq.Notification,
	traceID string,
) {
	logger.Log(
		"%s received '%s::%s' notification from %s (%d/i) [%s]",
		ref.ShortString(),
		n.Namespace,
		n.Type,
		n.ID.Ref.ShortString(),
		n.Payload.Len(),
		traceID,
	)
}

func logListen(
	logger twelf.Logger,
	ref ident.Ref,
	ns string,
) {
	logger.Debug(
		"%s started listening for notifications in '%s' namespace",
		ref.ShortString(),
		ns,
	)
}

func logUnlisten(
	logger twelf.Logger,
	ref ident.Ref,
	ns string,
) {
	logger.Debug(
		"%s stopped listening for notifications in '%s' namespace",
		ref.ShortString(),
		ns,
	)
}

func logSessionDestroy(
	logger twelf.Logger,
	ref ident.Ref,
	attrs attributes.Catalog,
	traceID string,
) {
	if traceID == "" {
		logger.Log(
			"%s session destroyed %s",
			ref.ShortString(),
			attrs,
		)
	} else {
		logger.Log(
			"%s session destroyed %s [%s]",
			ref.ShortString(),
			attrs,
			traceID,
		)
	}
}
