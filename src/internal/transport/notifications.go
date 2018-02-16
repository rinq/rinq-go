package transport

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/constraint"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// Notification is a transport-layer representation of an inter-session
// notification.
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

// Publisher is an interface for sending inter-session notifications.
type Publisher interface {
	// Publish sends a notification.
	Publish(*Notification) error
}

// Subscriber is an interface for receiving inter-session notifications.
type Subscriber interface {
	// Listen starts listening for notifications in the ns namespace.
	Listen(ns string) error

	// Unlisten stops listening for notifications in the ns namespace.
	// It must be called once for each prior call to Listen() before
	// notifications from ns are stopped.
	Unlisten(ns string) error

	// Consume accepts notifications and sends them to n until ctx is canceled.
	Consume(ctx context.Context, n chan<- *Notification) error

	// Ack acknowledges the notification with the given ID, it MUST be called
	// after each notification has been handled.
	Ack(ident.MessageID)
}
