package localsession

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

func traceCallBegin(
	span opentracing.Span,
	msgID ident.MessageID,
	ns string,
	cmd string,
	out *rinq.Payload,
) {
	traceCommandCommon(span, msgID, ns, cmd)
	traceCommandRequest(span, "rinq.call", out)
}

func traceCallEnd(
	span opentracing.Span,
	in *rinq.Payload,
	err error,
) {
	traceCommandResponse(span, in, err)
}

func traceAsyncRequest(
	span opentracing.Span,
	msgID ident.MessageID,
	ns string,
	cmd string,
	out *rinq.Payload,
	err error,
) {
	traceCommandCommon(span, msgID, ns, cmd)
	traceCommandRequest(span, "rinq.call-async", out)

	if err != nil {
		traceCommandError(span, err)
	}
}

func traceAsyncResponse(
	ctx context.Context,
	span opentracing.Span,
	msgID ident.MessageID,
	ns string,
	cmd string,
	in *rinq.Payload,
	err error,
) {
	traceCommandCommon(span, msgID, ns, cmd)
	traceCommandResponse(span, in, err)
}

func traceExecute(
	span opentracing.Span,
	msgID ident.MessageID,
	ns string,
	cmd string,
	out *rinq.Payload,
	err error,
	traceID string,
) {
	traceCommandCommon(span, msgID, ns, cmd)
	traceCommandRequest(span, "rinq.execute", out)
}

func traceCommandCommon(
	span opentracing.Span,
	msgID ident.MessageID,
	ns string,
	cmd string,
) {
	span.SetOperationName(ns + "::" + cmd)

	span.SetTag("rinq.id", msgID.String())
	span.SetTag("rinq.subsystem", "command")
	span.SetTag("rinq.namespace", ns)
	span.SetTag("rinq.command", cmd)
}

func traceCommandRequest(span opentracing.Span, event string, out *rinq.Payload) {
	span.LogFields(
		log.String("event", event),
		log.Int("rinq.payload_out", out.Len()),
	)
}

func traceCommandResponse(span opentracing.Span, in *rinq.Payload, err error) {
	if err == nil {
		traceCommandSuccess(span, in)
	} else {
		traceCommandError(span, err)
	}
}

func traceCommandSuccess(span opentracing.Span, in *rinq.Payload) {
	span.LogFields(
		log.String("event", "rinq.success"),
		log.Int("rinq.payload_in", in.Len()),
	)
}

func traceCommandError(span opentracing.Span, err error) {
	ext.Error.Set(span, true)

	switch e := err.(type) {
	case rinq.Failure:
		span.LogFields(
			log.String("event", "rinq.failure"),
			log.String("error.kind", e.Type),
			log.String("message", e.Message),
			log.Int("rinq.payload_in", e.Payload.Len()),
		)

	case rinq.CommandError:
		span.LogFields(
			log.String("event", "rinq.error"),
			log.Object("error.object", e),
		)

	default:
		span.LogFields(
			log.String("event", "error"),
			log.Object("error.object", err),
		)
	}
}

func traceNotifyCommon(
	span opentracing.Span,
	msgID ident.MessageID,
	ns string,
	t string,
) {
	span.SetOperationName(ns + "::" + t)

	span.SetTag("rinq.id", msgID.String())
	span.SetTag("rinq.subsystem", "notify")
	span.SetTag("rinq.namespace", ns)
	span.SetTag("rinq.type", t)
}

func traceNotifyUnicast(
	span opentracing.Span,
	msgID ident.MessageID,
	target ident.SessionID,
	ns string,
	t string,
	p *rinq.Payload,
) {
	traceNotifyCommon(span, msgID, ns, t)

	span.LogFields(
		log.String("event", "rinq.notify"),
		log.String("rinq.target", target.String()),
		log.Int("rinq.payload_out", p.Len()),
	)
}

func traceNotifyMulticast(
	span opentracing.Span,
	msgID ident.MessageID,
	con rinq.Constraint,
	ns string,
	t string,
	p *rinq.Payload,
) {
	traceNotifyCommon(span, msgID, ns, t)

	span.LogFields(
		log.String("event", "rinq.notify-many"),
		log.Object("rinq.constraint", con),
		log.Int("rinq.payload_out", p.Len()),
	)
}

func traceNotifyRecv(
	span opentracing.Span,
	msgID ident.MessageID,
	ref ident.Ref,
	n rinq.Notification,
) {
	traceNotifyCommon(span, msgID, n.Namespace, n.Type)

	fields := []log.Field{
		log.String("event", "rinq.notification"),
		log.String("rinq.target", ref.String()),
		log.Bool("rinq.is_multicast", n.IsMulticast),
	}

	if n.IsMulticast {
		fields = append(
			fields,
			log.Object("rinq.constraint", n.Constraint),
		)
	}

	fields = append(
		fields,
		log.Int("rinq.payload_in", n.Payload.Len()),
	)

	span.LogFields(fields...)
}
