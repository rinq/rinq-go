package rinq

import (
	"context"
	"fmt"

	"github.com/rinq/rinq-go/src/rinq/constraint"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// Session is an interface representing a "local" session, that is, a session
// created by a peer running in this process.
//
// Sessions are the "clients" on a Rinq network, able to issue command requests
// and send notifications to other sessions.
//
// Sessions are created by calling Peer.Session(). The peer that creates a
// session is called the "owning peer".
//
// Each session has an in-memory attribute table, which can be used to store
// application-defined key/value pairs. A session's attribute table can be
// modified locally, as well as remotely by peers that have received a command
// request or notification from the session.
//
// The attribute table is namespaced. Any operation performed on the attribute
// table occurs within a single namespace.
//
// The attribute table is versioned. Each revision of the attribute table is
// represented by the Revision interface.
//
// An optimistic-locking strategy is employed to protect the attribute table
// against concurrent writes. In order for a write to succeed, it must be made
// through a Revision value that represents the current (most recent) revision.
//
// Individual attributes in the table can be "frozen", preventing any further
// changes to that attribute.
type Session interface {
	// ID returns the session's unique identifier.
	ID() ident.SessionID

	// CurrentRevision returns the current revision of this session.
	//
	// If the session has been destroyed, any operation on the returned revision
	// will return a NotFoundError.
	CurrentRevision() Revision

	// Call sends a command request to the next available peer listening to the
	// ns namespace and waits for a response.
	//
	// In the context of the call, the sessions owning peer is the "client" and
	// the listening peer is the "server". The client and server may be the same
	// peer.
	//
	// cmd and out are an application-defined command name and request payload,
	// respectively. Both are passed to the command handler on the server.
	//
	// Calls always use a deadline; if ctx does not have a deadline, a timeout
	// described by options.DefaultTimeout() is used.
	//
	// If the call completes successfully, err is nil and in is the
	// application-defined response payload sent by the server.
	//
	// If err is non-nil, it may represent either client-side error or a
	// server-side error. IsServerError(err) returns true if the error occurred
	// on the server.
	//
	// IsFailure(err) returns true if the error is an application-defined
	// failure. Failures are server-side errors that are part of the command's
	// public API, as opposed to unexpected errors. If err is a failure, out
	// contains the failure's application-defined payload; for this reason
	// out.Close() must always be called, even if err is non-nil.
	//
	// If IsNotFound(err) returns true, the session has been destroyed and the
	// command request can not be sent.
	Call(ctx context.Context, ns, cmd string, out *Payload) (in *Payload, err error)

	// CallAync sends a command request to the next available peer listening to
	// the ns namespace and instructs it to send a response, but does not block.
	//
	// cmd and out are an application-defined command name and request payload,
	// respectively. Both are passed to the command handler on the server.
	//
	// id is a value identifying the outgoing command request.
	//
	// When a response is received, the handler specified by SetAsyncHandler()
	// is invoked. It is passed the id, namespace and command name of the
	// request, along with the response payload and error.
	//
	// It is the application's responsibility to correlate the request with the
	// response and handle the context deadline. The request is NOT tracked by
	// the session and as such the handler is never invoked in the event of a
	// timeout.
	//
	// If IsNotFound(err) returns true, the session has been destroyed and the
	// command request can not be sent.
	CallAsync(ctx context.Context, ns, cmd string, out *Payload) (id ident.MessageID, err error)

	// SetAsyncHandler sets the asynchronous call handler.
	//
	// h is invoked for each command response received to a command request made
	// with CallAsync().
	//
	// If IsNotFound(err) returns true, the session has been destroyed and the
	// command request can not be sent.
	SetAsyncHandler(h AsyncHandler) error

	// Execute sends a command request to the next available peer listening to
	// the ns namespace and returns immediately.
	//
	// cmd and out are an application-defined command name and request payload,
	// respectively. Both are passed to the command handler on the server.
	//
	// If IsNotFound(err) returns true, the session has been destroyed and the
	// command request can not be sent.
	Execute(ctx context.Context, ns, cmd string, out *Payload) (err error)

	// Notify sends a message directly to another session listening to the ns
	// namespace.
	//
	// t and out are an application-defined notification type and payload,
	// respectively. Both are passed to the notification handler configured on
	// the session identified by s.
	//
	// If IsNotFound(err) returns true, this session has been destroyed and the
	// notification can not be sent.
	Notify(ctx context.Context, ns, t string, s ident.SessionID, out *Payload) (err error)

	// NotifyMany sends a message to multiple sessions that are listening to the
	// ns namespace.
	//
	// The constraint c is a set of attribute key/value pairs that a session
	// must have in the ns namespace of its attribute table in order to receive
	// the notification.
	//
	// t and out are an application-defined notification type and payload,
	// respectively. Both are passed to the notification handlers configured on
	// those sessions that match c.
	//
	// If IsNotFound(err) returns true, this session has been destroyed and the
	// notification can not be sent.
	NotifyMany(ctx context.Context, ns, t string, c constraint.Constraint, out *Payload) error

	// Listen begins listening for notifications sent to this session in the ns
	// namespace.
	//
	// When a notification is received with a namespace equal to ns, h is invoked.
	//
	// h is invoked on its own goroutine for each notification.
	Listen(ns string, h NotificationHandler) error

	// Unlisten stops listening for notifications from the ns namespace.
	//
	// If the session is not currently listening for notifications, nil is
	// returned immediately.
	Unlisten(ns string) error

	// Destroy terminates the session.
	//
	// Destroy does NOT block until the session is destroyed, use the
	// Session.Done() channel to wait for the session to be destroyed.
	Destroy()

	// Done returns a channel that is closed when the session is destroyed and
	// any pending Session.Call() operations have completed.
	//
	// The session may be destroyed directly with Destroy(), or via a Revision
	// that refers to this session, either locally or remotely.
	//
	// All sessions are destroyed when their owning peer is stopped.
	Done() <-chan struct{}
}

// AsyncHandler is a call-back function invoked when a response is received to
// a command call made with Session.CallAsync()
//
// If err is non-nil, it always represents a server-side error.
//
// IsFailure(err) returns true if the error is an application-defined
// failure. Failures are server-side errors that are part of the command's
// public API, as opposed to unexpected errors. If err is a failure, in
// contains the failure's application-defined payload; for this reason
// in.Close() must be called, even if err is non-nil.
//
// The handler is responsible for closing the in payload, however there is no
// requirement that the payload be closed during the execution of the handler.
type AsyncHandler func(
	ctx context.Context,
	sess Session, msgID ident.MessageID,
	ns, cmd string,
	in *Payload, err error,
)

// NotFoundError indicates that an operation failed because the session does
// not exist.
type NotFoundError struct {
	ID ident.SessionID
}

// IsNotFound returns true if err is a NotFoundError.
func IsNotFound(err error) bool {
	_, ok := err.(NotFoundError)
	return ok
}

func (err NotFoundError) Error() string {
	return fmt.Sprintf("session %s not found", err.ID)
}
