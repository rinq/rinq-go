package overpass

import "context"

// Session is an interface for sending command requests and notifications on an
// Overpass network.
//
// Each sessions has an attribute table that stores application-defined
// key/value pairs. Multiple attributes can be modified in a single update
// operation, each successful update produces a new session "revision".
type Session interface {
	// ID returns the session's unique identifier.
	ID() SessionID

	// CurrentRevision returns the current revision of this session.
	CurrentRevision() (Revision, error)

	// Call sends a command request to one of the peers listening to the
	// specified namespace and blocks until a response is received.
	//
	// IsFailure(err) returns true if the error represents an application-defined
	// command failure sent from the peer that handled the command.
	//
	// The returned payload must be closed regardless of the error value. When
	// IsFailure(err) is true the payload returned is the payload from the
	// failure error value.
	Call(ctx context.Context, ns, cmd string, p *Payload) (*Payload, error)

	// Execute sends a command request to one of the peers listening to the
	// specified namespace, without waiting for a response.
	Execute(ctx context.Context, ns, cmd string, p *Payload) error

	// ExecuteMany sends a command request to all of the peers currently listening
	// to the specified namespace, without waiting for a response.
	ExecuteMany(ctx context.Context, ns, cmd string, p *Payload) error

	// Notify sends a message directly to a another session.
	Notify(ctx context.Context, target SessionID, typ string, p *Payload) error

	// NotifyMany sends a message to all sessions that have a specific set of
	// attributes.
	NotifyMany(ctx context.Context, con Constraint, typ string, p *Payload) error

	// Listen begins listening for notifications sent to this session.
	Listen(handler NotificationHandler) error

	// Unlisten stops listening for notifications.
	Unlisten() error

	// Close immediately terminates the session.
	Close()

	// Done returns a channel that is closed when the session is closed.
	Done() <-chan struct{}
}
