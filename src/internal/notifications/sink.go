package notifications

import "context"

// Sink is an interface that accepts notifications.
type Sink interface {
	// Send publishes a notification.
	Send(context.Context, *Notification) error
}
