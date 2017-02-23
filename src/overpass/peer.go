package overpass

// Peer represents a connection to an Overpass network.
//
// Any given peer can operate as a source for sessions (client-like behaviour)
// or a command handler (server-like behaviour), or both.
type Peer interface {
	// ID returns the peer's unique identifier.
	ID() PeerID

	// Session returns a new session belonging to this peer.
	Session() Session

	// Listen starts listening for command requests in the given namespace.
	Listen(namespace string, handler CommandHandler) error

	// Unlisten stops listening for command requests in the given namepsace.
	Unlisten(namespace string) error

	// Done returns a channel that is closed when the peer is closed.
	Done() <-chan struct{}

	// Err returns the error that caused the Done() channel to close, if any.
	Err() error

	// Stop disconnects the peer from the network.
	Stop() error

	// GracefulStop() disconnects the peer from the network after any pending
	// calls have returned, and pending command requests have been handled.
	GracefulStop() error
}
