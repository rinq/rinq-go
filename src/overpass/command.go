package overpass

import (
	"context"
	"errors"
	"fmt"
	"regexp"
)

// Command represents an incomming command request.
type Command struct {
	// Source refers to the session that sent the request.
	Source Revision

	// Namespace holds the command namespace, each namespace has its own
	// message queue.
	Namespace string

	// Command is the command to be executed.
	Command string

	// Payload contains optional application-defined information about the
	// request, such as arguments to the command.
	Payload *Payload

	// IsMulticast is true if the command request was (potentially) sent to more
	// than one peer.
	IsMulticast bool
}

// CommandHandler is a callback-function invoked when a command request is
// received by the peer.
//
// The handler MUST close the responder by calling Done(), Error() or Close();
// otherwise the command may be redelivered, possibly to a different peer.
type CommandHandler func(
	context.Context,
	Command,
	Responder,
)

// Responder is used to send responses to incoming command requests.
type Responder interface {
	// IsRequired returns true if the source session is waiting for the response.
	//
	// If the response is not required, any payload data sent is discarded.
	// The responder must always be closed, even if IsRequired() returns false.
	IsRequired() bool

	// IsClosed true if the responder has been closed.
	IsClosed() bool

	// Done sends a payload to the source session and closes the responder.
	//
	// A panic occurs if the responder is already closed.
	Done(*Payload)

	// Error sends an error to the source session and closes the responder.
	//
	// A panic occurs if the responder is already closed.
	Error(error)

	// Fail is a convenience method that creates a Failure and passes it to
	// the responder's Error() method.
	Fail(failureType, message string)

	// Close marks the responder as closed.
	//
	// If the origin session is waiting for a response it will receive a nil
	// payload.
	//
	// The response instance can not be used after it is closed.
	// It is not an error to close a responder multiple times.
	Close()
}

// IsValidNamespace returns an error if the given namespace is invalid.
func IsValidNamespace(namespace string) error {
	if namespace == "" {
		return errors.New("namespace must not be empty")
	} else if namespace[0] == '_' {
		return fmt.Errorf("namespace '%s' is reserved", namespace)
	} else if !namespacePattern.MatchString(namespace) {
		return fmt.Errorf("namespace '%s' contains invalid characters", namespace)
	}

	return nil
}

var namespacePattern *regexp.Regexp

func init() {
	namespacePattern = regexp.MustCompile(`^[A-Za-z0-9_\.\-:]+$`)
}
