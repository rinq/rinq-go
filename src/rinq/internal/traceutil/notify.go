package traceutil

import (
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
)

var (
	notifierUnicastEvent   = log.String("event", "notify")
	notifierMulticastEvent = log.String("event", "notify-many")
	listenerReceiveEvent   = log.String("event", "notification")
)

// SetupNotification configures span as a command-related span.
func SetupNotification(
	s opentracing.Span,
	id ident.MessageID,
	ns string,
	t string,
) {
	s.SetOperationName(ns + "::" + t + " notification")

	s.SetTag("subsystem", "notify")
	s.SetTag("message_id", id.String())
	s.SetTag("namespace", ns)
	s.SetTag("type", t)
}

// LogNotifierUnicast logs information about a unicast notification to s.
func LogNotifierUnicast(
	s opentracing.Span,
	attrs attrmeta.NamespacedTable,
	target ident.SessionID,
	p *rinq.Payload,
) {
	fields := []log.Field{
		notifierUnicastEvent,
		log.String("target", target.String()),
		log.Int("size", p.Len()),
	}

	if len(attrs) > 0 {
		fields = append(fields, lazyString("attributes", attrs.String))
	}

	s.LogFields(fields...)
}

// LogNotifierMulticast logs informatin about a multicast notification to s.
func LogNotifierMulticast(
	s opentracing.Span,
	attrs attrmeta.NamespacedTable,
	con rinq.Constraint,
	p *rinq.Payload,
) {
	fields := []log.Field{
		notifierMulticastEvent,
		log.String("constraint", con.String()),
		log.Int("size", p.Len()),
	}

	if len(attrs) > 0 {
		fields = append(fields, lazyString("attributes", attrs.String))
	}

	s.LogFields(fields...)
}

// LogNotifierError logs information about err to s.
func LogNotifierError(s opentracing.Span, err error) {
	ext.Error.Set(s, true)

	s.LogFields(
		errorEvent,
		log.String("message", err.Error()),
	)
}

// LogListenerReceived logs information about a received notification to s.
func LogListenerReceived(s opentracing.Span, ref ident.Ref, n rinq.Notification) {
	fields := []log.Field{
		listenerReceiveEvent,
		log.String("recipient", ref.String()),
		log.Bool("multicast", n.IsMulticast),
		log.Int("size", n.Payload.Len()),
	}

	if n.IsMulticast {
		fields = append(
			fields,
			log.String("constraint", n.Constraint.String()),
		)
	}

	s.LogFields(fields...)
}
