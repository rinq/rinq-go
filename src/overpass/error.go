package overpass

import "fmt"

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
