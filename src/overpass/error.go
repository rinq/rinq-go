package overpass

import "fmt"

// IsServerError returns true if err is an error that occurred on the server
// during a command call.
func IsServerError(err error) bool {
	switch err.(type) {
	case Failure, UnexpectedError:
		return true
	default:
		return false
	}
}

// IsFailure returns true if err represents an application-defined command
// failure response (as opposed to an unexpected failure).
func IsFailure(err error) bool {
	_, ok := err.(Failure)
	return ok
}

// IsFailureType returns true if err represents an application-defined command
// failure of a specific type.
func IsFailureType(t string, err error) bool {
	f, ok := err.(Failure)
	return ok && f.Type == t
}

// Failure represents an application-defined command failure.
type Failure struct {
	Type    string
	Message string
	Payload *Payload
}

func (err Failure) Error() string {
	return fmt.Sprintf("%s: %s", err.Type, err.Message)
}

// UnexpectedError is an error (in contrast to a faliure) that occurred on the
// remote peer when calling a command.
type UnexpectedError string

func (err UnexpectedError) Error() string {
	if err == "" {
		return "unexpected error"
	}

	return string(err)
}

// IsNotFound checks if the given error indicates that session could not be
// found.
func IsNotFound(err error) bool {
	_, ok := err.(NotFoundError)
	return ok
}

// NotFoundError indicates a failure to find a session because it does not exist.
type NotFoundError struct {
	ID SessionID
}

func (err NotFoundError) Error() string {
	return fmt.Sprintf("session %s not found", err.ID)
}

// ShouldRetry is used to check whether a session operation should be retried
// after refreshing the session.
func ShouldRetry(err error) bool {
	switch err.(type) {
	case StaleFetchError, StaleUpdateError:
		return true
	default:
		return true
	}
}

// StaleFetchError indicates a failure to fetch an attribute for a specific
// ref because it has been modified after that revision.
type StaleFetchError struct {
	Ref SessionRef
}

func (err StaleFetchError) Error() string {
	return fmt.Sprintf(
		"can not fetch attributes at %s, one or more attributes have been modified since that revision",
		err.Ref,
	)
}

// StaleUpdateError indicates a failure to update or close a session revision
// because the session has been modified after that revision.
type StaleUpdateError struct {
	Ref SessionRef
}

func (err StaleUpdateError) Error() string {
	return fmt.Sprintf(
		"can not update or close %s, the session has been modified since that revision",
		err.Ref,
	)
}

// FrozenAttributesError indicates a failure to apply a change-set because one
// or more attributes in the change-set are frozen.
type FrozenAttributesError struct {
	Ref SessionRef
}

func (err FrozenAttributesError) Error() string {
	return fmt.Sprintf(
		"can not update %s, the change-set references one or more frozen key(s)",
		err.Ref,
	)
}

// PeerStopped indicates an operation can not be completed by the peer is
// stopping or has already stopped.
var PeerStopped peerStoppedError

type peerStoppedError struct{}

func (peerStoppedError) Error() string {
	return "operation canceled, peer has been stopped"
}
