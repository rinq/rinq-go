package overpass

import "fmt"

// Failure is an application-defined command error.
//
// Failures are used to indicate an error that is "expected" within the domain
// of the command that produced it. The for part of the command's API and should
// usually be handled by the caller.
//
// Failures can be produced by a command handler by calling Response.Fail() or
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
