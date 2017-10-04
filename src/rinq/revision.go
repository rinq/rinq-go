package rinq

import (
	"context"
	"fmt"

	"github.com/rinq/rinq-go/src/rinq/ident"
)

// Revision represents a specific revision of a session.
//
// Revision is the sole interface for manipulating a session's attribute table.
//
// The underlying session may be "local", i.e. owned by a peer running in this
// process, or "remote", owned by a different peer.
//
// For remote sessions, operations may require network IO. Deadlines are
// honored for all methods that accept a context.
type Revision interface {
	// Ref returns the session reference, which uniquely identifies the session
	// ID and the revision number represented by this instance.
	Ref() ident.Ref

	// Refresh returns the latest revision of the session.
	//
	// If IsNotFound(err) returns true, the session has been destroyed and rev is
	// invalid.
	Refresh(ctx context.Context) (rev Revision, err error)

	// Get returns the attribute with key k within the ns namespace of the
	// attribute table.
	//
	// The returned attribute is guaranteed to be correct as of Ref().Rev.
	// Non-existent attributes are equivalent to empty attributes, therefore it
	// is not an error to request a key that has never been created.
	//
	// Peers do not always have a complete copy of the attribute table. If the
	// attribute value is unknown it is fetched from the owning peer.
	//
	// If the attribute can not be retrieved because it has already been
	// modified, ShouldRetry(err) returns true. To fetch the attribute value at
	// the later revision, first call Refresh() then retry the Get() on the
	// newer revision.
	//
	// If IsNotFound(err) returns true, the session has been destroyed and the
	// revision can not be queried.
	Get(ctx context.Context, ns, k string) (attr Attr, err error)

	// GetMany returns the attributes with keys in k within the ns namespace of
	// the attribute table.
	//
	// The returned attributes are guaranteed to be correct as of Ref().Rev.
	// Non-existent attributes are equivalent to empty attributes, therefore it
	// is not an error to request keys that have never been created.
	//
	// Peers do not always have a complete copy of the attribute table. If any
	// of the attribute values are unknown they are fetched from the owning peer.
	//
	// If any of the attributes can not be retrieved because they have already
	// been modified, ShouldRetry(err) returns true. To fetch the attribute
	// values at the later revision, first call Refresh() then retry the
	// GetMany() on the newer revision.
	//
	// If IsNotFound(err) returns true, the session has been destroyed and the
	// revision can not be queried.
	//
	// If err is nil, t contains all of the attributes specified in k.
	GetMany(ctx context.Context, ns string, k ...string) (t AttrTable, err error)

	// Update atomically modifies a set of attributes within the ns namespace of
	// the attribute table.
	//
	// A successful update produces a new revision.
	//
	// Each update is atomic; either all of the attributes in attrs are updated,
	// or the attribute table remains unchanged. On success, rev is the newly
	// created revision.
	//
	// The following conditions must be met for an update to succeed:
	//
	// 1. The session revision represented by this instance must be the latest
	//    revision. If Ref().Rev is not the latest revision the update fails;
	//    ShouldRetry(err) returns true.
	//
	// 2. All attribute changes must reference non-frozen attributes. If any of
	//    attributes being updated are already frozen the update fails and
	//    ShouldRetry(err) returns false.
	//
	// If attrs is empty no update occurs, rev is this revision and err is nil.
	//
	// As a convenience, if the update fails for any reason, rev is this
	// revision. This allows the caller to assign the return value to an
	// existing variable without first checking for errors.
	Update(ctx context.Context, ns string, attrs ...Attr) (rev Revision, err error)

	// Clear is an update operation that atomically sets the value of each
	// attribute within the ns namespace to the empty string.
	//
	// The sematics are the same as for Update(). This means the operation fails
	// if ANY attribute in the ns namespace is frozen.
	//
	// As a convenience, if the clear operation fails for any reason, rev is
	// this revision. This allows the caller to assign the return value to an
	// existing variable without first checking for errors.
	Clear(ctx context.Context, ns string) (rev Revision, err error)

	// Destroy terminates the session.
	//
	// The session revision represented by this instance must be the latest
	// revision. If Ref().Rev is not the latest revision the destroy fails;
	// ShouldRetry(err) returns true.
	Destroy(ctx context.Context) (err error)
}

// ShouldRetry returns true if a call to Revision.Get(), GetMany(), Update() or
// Destroy() failed because the revision is out of date.
//
// The operation should be retried on the latest revision of the session,
// which can be retrieved with Revision.Refresh().
func ShouldRetry(err error) bool {
	switch err.(type) {
	case StaleFetchError, StaleUpdateError:
		return true
	default:
		return false
	}
}

// StaleFetchError indicates a failure to fetch an attribute for a specific
// revision because it has been modified after that revision.
type StaleFetchError struct {
	Ref ident.Ref
}

func (err StaleFetchError) Error() string {
	return fmt.Sprintf(
		"can not fetch attributes at %s, one or more attributes have been modified since that revision",
		err.Ref,
	)
}

// StaleUpdateError indicates a failure to update or destroy a session because
// the session has been modified after that revision.
type StaleUpdateError struct {
	Ref ident.Ref
}

func (err StaleUpdateError) Error() string {
	return fmt.Sprintf(
		"can not update or close %s, the session has been modified since that revision",
		err.Ref,
	)
}

// FrozenAttributesError indicates a failure to update a session because at least
// one of the attributes being updated is frozen.
type FrozenAttributesError struct {
	Ref ident.Ref
}

func (err FrozenAttributesError) Error() string {
	return fmt.Sprintf(
		"can not update %s, the change-set references one or more frozen key(s)",
		err.Ref,
	)
}
