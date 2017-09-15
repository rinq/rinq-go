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

	// Payload contains optional application-defined information.
	Payload *Payload

	// IsMulticast is true if the notification was (potentially) sent to more
	// than one session.
	IsMulticast bool

	// For multicast notifications, Constraint contains the attributes used as
	// criteria for selecting which sessions receive the notification.
	Constraint Constraint
}

// NotificationHandler is a callback-function invoked when a inter-session
// notification is received.
type NotificationHandler func(
	ctx context.Context,
	target Session,
	n Notification,
)
