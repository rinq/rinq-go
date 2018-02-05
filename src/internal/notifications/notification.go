package notifications

import (
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/constraint"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// Common contains the fields of notifications that are shared by inbound and
// outbound notifications.
type Common struct {
	ID                  ident.MessageID
	TraceID             string
	Namespace           string
	Type                string
	Payload             *rinq.Payload
	IsMulticast         bool
	UnicastTarget       ident.SessionID
	MulticastConstraint constraint.Constraint
}

// Outbound is a notification sent to a sink.
type Outbound struct {
	Common

	Span opentracing.Span
}

// Inbound is a notification received from a source.
type Inbound struct {
	Common

	SpanContext opentracing.SpanContext
	Ack         func()
}
