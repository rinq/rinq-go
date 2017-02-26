package overpass

import "fmt"

// Failure represents an application-defined command failure.
//
// Failures are returned by commands to indicate an error that is "expected"
// within the domain of the command. Failures form part of a command's public
// API and should usually be handled by the client.
type Failure struct {
	// Type is an application-defined failure type. There is no restriction on
	// the content, but the type must not be empty. Overpass logs the failure
	// type for each failed request.
	Type string

	// Message is an optional human-readable description of the failure.
	Message string

	// Payload is an optional application-defined payload sent along with the
	// failure.
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

// FailureType returns the type of err if is a Failure; otherwise is returns an
// empty string.
func FailureType(err error) string {
	f, ok := err.(Failure)
	if ok && f.Type == "" {
		panic("failure type is empty")
	}

	return f.Type
}
