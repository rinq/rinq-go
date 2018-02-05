package notifications

// Source is an interface that produces notifications.
type Source interface {
	// Unlisten starts listening for notifications in the ns namespace.
	Listen(ns string)

	// Unlisten stops listening for notifications in the ns namespace.
	Unlisten(ns string)

	// Messages returns a channel on which incoming notifications are delivered.
	Notifications() <-chan *Inbound
}
