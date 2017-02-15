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

	// Wait blocks until the peer is closed or an error occurs.
	Wait() error

	// Close disconnects the peer from the network.
	Close()
}
