package transport

import (
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

// InboundNotification is a notification received from a consumer.
// Done() must be called once the notification has been handled by all target
// sessions.
type InboundNotification struct {
	*Notification
	Done func()
}

// Publisher is an interface for sending inter-session notifications.
type Publisher interface {
	// Publish sends a notification.
	Publish(*Notification) error
}

// Subscriber is an interface for subscribing to and unsubscribing from
// notifications on a per-namespace basis.
type Subscriber interface {
	// Listen starts listening for notifications in the ns namespace.
	Listen(ns string) error

	// Unlisten stops listening for notifications in the ns namespace.
	// It must be called once for each prior call to Listen() before
	// notifications from ns are stopped.
	Unlisten(ns string) error
}

// Consumer is an interface for receiving inter-session notifications.
type Consumer interface {
	// Queue returns a channel on which inbound notifications are delivered.
	Queue() <-chan InboundNotification
}
