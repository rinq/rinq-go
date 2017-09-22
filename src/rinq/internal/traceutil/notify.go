package traceutil

import (
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

const (
	typeKey       = "type"
	targetKey     = "target"
	constraintKey = "constraint"
	multicastKey  = "multicast"
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
	s.SetOperationName(ns + "::" + t)

	s.SetTag(subsystemKey, "notify")
	s.SetTag(messageIDKey, id.String())
	s.SetTag(namespaceKey, ns)
	s.SetTag(typeKey, t)
}

// LogNotifierUnicast logs information about a unicast notification to s.
func LogNotifierUnicast(
	s opentracing.Span,
	target ident.SessionID,
	p *rinq.Payload,
) {
	s.LogFields(
		notifierUnicastEvent,
		log.String(targetKey, target.String()),
		log.Int(payloadSizeKey, p.Len()),
	)
}

// LogNotifierMulticast logs informatin about a multicast notification to s.
func LogNotifierMulticast(
	s opentracing.Span,
	con rinq.Constraint,
	p *rinq.Payload,
) {
	s.LogFields(
		notifierMulticastEvent,
		log.String(constraintKey, con.String()),
		log.Int(payloadSizeKey, p.Len()),
	)
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
	s.SetTag(targetKey, ref.String())

	fields := []log.Field{
		listenerReceiveEvent,
		log.Bool(multicastKey, n.IsMulticast),
		log.Int(payloadSizeKey, n.Payload.Len()),
	}

	if n.IsMulticast {
		fields = append(
			fields,
			log.String(constraintKey, n.Constraint.String()),
		)
	}

	s.LogFields(fields...)
}
