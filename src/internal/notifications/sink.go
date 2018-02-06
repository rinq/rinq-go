package notifications

// Sink is an interface that accepts notifications.
type Sink interface {
	// Send publishes a notification.
	Send(*Notification) error
}
