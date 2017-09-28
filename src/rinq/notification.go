package rinq

import "context"

// Notification holds information about an inter-session notification.
type Notification struct {
	// Source refers to the session that sent the notification.
	Source Revision

	// Namespace is the notification namespace. Namespaces are used to route
	// notifications to only those sessions that intend to handle them.
	Namespace string

	// Type is an application-defined notification type.
	Type string

	// Payload contains optional application-defined information. The handler
	// that accepts the notifiation is responsible for closing the payload,
	// however there is no requirement that the payload be closed during the
	// execution of the handler.
	Payload *Payload

	// IsMulticast is true if the notification was (potentially) sent to more
	// than one session.
	IsMulticast bool

	// For multicast notifications, Constraint contains the attributes used as
	// criteria for selecting which sessions receive the notification. The
	// constraint is nil if IsMulticast is false.
	Constraint Constraint
}

// NotificationHandler is a callback-function invoked when an inter-session
// notification is received.
//
// Notifications can only be received for namespaces that a session is listening
// to. See Session.Listen() to start listening.
//
// The handler is responsible for closing n.Payload, however there is no
// requirement that the payload be closed during the execution of the handler.
type NotificationHandler func(
	ctx context.Context,
	target Session,
	n Notification,
)
