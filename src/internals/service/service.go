package service

// Service is an interface for background tasks that can finish with an error.
type Service interface {
	// Done returns a channel that is closed when the session is closed.
	Done() <-chan struct{}

	// Err returns the error that caused the Done() channel to close, if any.
	Err() error

	// Stop halts the service immediately.
	Stop() error
}
