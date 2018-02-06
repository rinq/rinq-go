package notifications

// Source is a channel that supplies incoming notifications.
type Source <-chan Inbound

// Inbound is a notification received from a source.
// Ack must be called once the notification has been handled.
type Inbound struct {
	Notification *Notification
	Ack          func()
}
