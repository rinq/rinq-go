package overpass

import (
	"context"
	"errors"
	"fmt"
	"regexp"
)

// CommandHandler is a callback-function invoked when a command request is
// received by the peer.
//
// Command requests can only be received for namespaces that a peer is listening
// to. See Peer.Listen() to start listening.
//
// The handler MUST close the response by calling r.Done(), r.Error() or
// r.Close(); otherwise the request may be redelivered, possibly to a different
// peer.
type CommandHandler func(
	ctx context.Context,
	q Request,
	r Response,
)

// Request holds information about an incoming command request.
type Request struct {
	// Source is the revision of the session that sent the request, at the time
	// it was sent (which is not necessarily the latest).
	Source Revision

	// Namespace is the command namespace. Namespaces are used to route command
	// requests to the appropriate peer and comand handler.
	Namespace string

	// Command is the application-defined command name for the request. Overpass
	// logs the command name for each request.
	Command string

	// Payload contains optional application-defined information about the
	// request, such as arguments to the command.
	Payload *Payload

	// IsMulticast is true if the command request was (potentially) sent to more
	// than one peer using Session.ExecuteMany().
	IsMulticast bool
}

// Response sends a reply to incoming command requests.
type Response interface {
	// IsRequired returns true if the source session is waiting for the response.
	//
	// If the response is not required, any payload data sent is discarded.
	// The response must always be closed, even if IsRequired() returns false.
	IsRequired() bool

	// IsClosed true if the response has already been closed.
	IsClosed() bool

	// Done sends a payload to the source session and closes the response.
	//
	// A panic occurs if the response has already been closed.
	Done(*Payload)

	// Error sends an error to the source session and closes the response.
	//
	// A panic occurs if the response has already been closed.
	Error(error)

	// Fail is a convenience method that creates a Failure and passes it to
	// Error() method. The created failure is returned.
	//
	// A panic occurs if the response has already been closed.
	Fail(failureType, message string) Failure

	// Close finalizes the response.
	//
	// If the origin session is expecting response it will receive a nil payload.
	//
	// It is not an error to close a responder multiple times. The return value
	// is true the first time Close() is called, and false on subsequent calls.
	Close() bool
}

// ValidateNamespace checks if ns is a valid namespace.
//
// Namespaces can contain alpha-numeric characters, underscores, hyphens,
// periods and colons.
//
// Namespaces beginning with an underscore are reserved for internal use.
//
// The return value is nil if ns is valid, unreserved namespace.
func ValidateNamespace(ns string) error {
	if ns == "" {
		return errors.New("namespace must not be empty")
	} else if ns[0] == '_' {
		return fmt.Errorf("namespace '%s' is reserved", ns)
	} else if !namespacePattern.MatchString(ns) {
		return fmt.Errorf("namespace '%s' contains invalid characters", ns)
	}

	return nil
}

var namespacePattern *regexp.Regexp

func init() {
	namespacePattern = regexp.MustCompile(`^[A-Za-z0-9_\.\-:]+$`)
}
