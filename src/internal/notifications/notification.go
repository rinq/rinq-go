package notifications

import (
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/constraint"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// Notification is a low-level representation of a notification.
type Notification struct {
	ID                  ident.MessageID
	TraceID             string
	SpanContext         opentracing.SpanContext
	Namespace           string
	Type                string
	Payload             *rinq.Payload
	IsMulticast         bool
	UnicastTarget       ident.SessionID
	MulticastConstraint constraint.Constraint
}
