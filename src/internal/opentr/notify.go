package opentr

import (
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/rinq/rinq-go/src/internal/attributes"
	"github.com/rinq/rinq-go/src/internal/transport"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

var (
	notifierUnicastEvent   = log.String("event", "notify")
	notifierMulticastEvent = log.String("event", "notify-many")
	listenerReceiveEvent   = log.String("event", "notification")
)

// notification configures s as a notification-related span.
func notification(
	t opentracing.Tracer,
	n *transport.Notification,
	opts []opentracing.StartSpanOption,
) opentracing.Span {
	opts = append(
		CommonSpanOptions,
		opts...,
	)

	s := t.StartSpan(
		n.Namespace+"::"+n.Type+" notification",
		opts...,
	)

	s.SetTag("subsystem", "notify")
	s.SetTag("message_id", n.ID.String())
	s.SetTag("namespace", n.Namespace)
	s.SetTag("type", n.Type)
	s.SetTag("trace_id", n.TraceID)

	return s
}

// SentNotification returns a span representing an outbound notification.
// attrs is the session's attributes at the time the notification was sent.
func SentNotification(
	t opentracing.Tracer,
	n *transport.Notification,
	attrs attributes.Catalog,
) opentracing.Span {
	opts := []opentracing.StartSpanOption{
		ext.SpanKindProducer,
	}

	if n.SpanContext != nil {
		opts = append(
			opts,
			opentracing.FollowsFrom(n.SpanContext),
		)
	}

	s := notification(t, n, opts)

	var fields []log.Field

	if n.IsMulticast {
		fields = append(
			fields,
			notifierMulticastEvent,
			log.String("constraint", n.MulticastConstraint.String()),
		)
	} else {
		fields = append(
			fields,
			notifierUnicastEvent,
			log.String("target", n.UnicastTarget.String()),
		)
	}

	fields = append(
		fields,
		log.Int("size", n.Payload.Len()),
	)

	if len(attrs) > 0 {
		fields = append(fields, lazyString("attributes", attrs.String))
	}

	s.LogFields(fields...)

	return s
}

// ReceivedNotification returns a span representing an inbound notification.
// ref is the session's ref at the time the notification was received.
// n.SpanContext is set to the span context for the returned span.
func ReceivedNotification(
	t opentracing.Tracer,
	ref ident.Ref,
	n *transport.Notification,
) opentracing.Span {
	opts := []opentracing.StartSpanOption{
		ext.SpanKindConsumer,
	}
	s := notification(t, n, opts)

	n.SpanContext = s.Context()

	fields := []log.Field{
		listenerReceiveEvent,
		log.String("recipient", ref.String()),
		log.Bool("multicast", n.IsMulticast),
		log.Int("size", n.Payload.Len()),
	}

	if n.IsMulticast {
		fields = append(
			fields,
			log.String("constraint", n.MulticastConstraint.String()),
		)
	}

	s.LogFields(fields...)

	return s
}

// LogNotificationError logs information about err to s.
func LogNotificationError(s opentracing.Span, err error) {
	ext.Error.Set(s, true)

	s.LogFields(
		errorEvent,
		log.String("message", err.Error()),
	)
}
