package rinq

import (
	"context"
	"fmt"

	"github.com/rinq/rinq-go/src/rinq/ident"
)

// CommandHandler is a callback-function invoked when a command request is
// received by the peer.
//
// Command requests can only be received for namespaces that a peer is listening
// to. See Peer.Listen() to start listening.
//
// The handler MUST close the response by calling res.Done(), res.Error() or
// res.Close(); otherwise the request may be redelivered, possibly to a
// different peer.
//
// The handler is responsible for closing req.Payload, however there is no
// requirement that the payload be closed during the execution of the handler.
type CommandHandler func(
	ctx context.Context,
	req Request,
	res Response,
)

// Request holds information about an incoming command request.
type Request struct {
	// ID uniquely identifies the command request.
	ID ident.MessageID

	// Source is the revision of the session that sent the request, at the time
	// it was sent (which is not necessarily the latest).
	Source Revision

	// Namespace is the command namespace. Namespaces are used to route command
	// requests to the appropriate peer and command handler.
	Namespace string

	// Command is the application-defined command name for the request. The
	// command is logged for each request.
	Command string

	// Payload contains optional application-defined information about the
	// request, such as arguments to the command. The handler that accepts the
	// request is responsible for closing the payload, however there is no
	// requirement that the payload be closed during the execution of the handler.
	Payload *Payload
}

// Response sends a reply to incoming command requests.
type Response interface {
	// IsRequired returns true if the sender is waiting for the response.
	//
	// If the response is not required, any payload data sent is discarded.
	// The response must always be closed, even if IsRequired() returns false.
	IsRequired() bool

	// IsClosed returns true if the response has already been closed.
	IsClosed() bool

	// Done sends a payload to the source session and closes the response.
	//
	// A panic occurs if the response has already been closed.
	Done(*Payload)

	// Error sends an error to the source session and closes the response.
	//
	// A panic occurs if the response has already been closed.
	Error(error)

	// Fail is a convenience method that creates a Failure and passes it to the
	// Error() method. The created failure is returned.
	//
	// The failure type t is used verbatim. The failure message is formatted
	// according to the format specifier f, interpolated with values from v.
	//
	// A panic occurs if the response has already been closed or if t is empty.
	Fail(t, f string, v ...interface{}) Failure

	// Close finalizes the response.
	//
	// If the origin session is expecting response it will receive a nil payload.
	//
	// It is not an error to close a response multiple times. The return value
	// is true the first time Close() is called, and false on subsequent calls.
	Close() bool
}

// Failure is an application-defined command error.
//
// Failures are used to indicate an error that is "expected" within the domain
// of the command that produced it. Failures form part of the command's API and
// should usually be handled by the caller.
//
// Failures can be produced in a command handler by calling Response.Fail() or
// passing a Failure value to Response.Error().
type Failure struct {
	// Type is an application-defined string identifying the failure.
	// They serve the same purpose as an error code. They should be concise
	// and easily understandable within the context of the application's API.
	Type string

	// Message is an optional human-readable description of the failure.
	Message string

	// Payload is an optional application-defined payload.
	Payload *Payload
}

func (err Failure) Error() string {
	return fmt.Sprintf("%s: %s", err.Type, err.Message)
}

// IsFailure returns true if err is a Failure.
func IsFailure(err error) bool {
	_, ok := err.(Failure)
	return ok
}

// IsFailureType returns true if err is a Failure with a type of t.
func IsFailureType(t string, err error) bool {
	if t == "" {
		panic("failure type is empty")
	}

	f, _ := err.(Failure)
	return f.Type == t
}

// FailureType returns the failure type of err; or an empty string if err is not
// a Failure.
func FailureType(err error) string {
	f, ok := err.(Failure)
	if ok && f.Type == "" {
		panic("failure type is empty")
	}

	return f.Type
}

// IsCommandError returns true if err was sent in response to a command request,
// as opposed to a local error that occurred when attempting to send the request.
func IsCommandError(err error) bool {
	switch err.(type) {
	case Failure, CommandError:
		return true
	default:
		return false
	}
}

// CommandError is an error (as opposed to a Failure) sent in response to a
// command.
type CommandError string

func (err CommandError) Error() string {
	if err == "" {
		return "unexpected command error"
	}

	return string(err)
}
