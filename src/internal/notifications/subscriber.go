package notifications

// Subscriber is an interface for controlling which namespaces to monitor for
// notifications.
//
// It is used in conjunction with a Source to provide notifications to a
// Dispatcher.
type Subscriber interface {
	// Listen starts listening for notifications in the ns namespace.
	Listen(ns string) error

	// Unlisten stops listening for notifications in the ns namespace.
	// It must be called once for each prior call to Listen() before
	// notifications from ns are stopped.
	Unlisten(ns string) error
}
