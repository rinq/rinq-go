package overpass

import "context"

// Revision represents a specific revision of session.
//
// Revision is the sole interface for manipulating a session's attribute table.
//
// The underlying session may be "local", i.e. owned by a peer running in this
// process, or "remote", owned by a different peer.
//
// For remote sessions, operations may require network IO. Deadlines are
// honoured for all methods that accept a context.
type Revision interface {
	// Ref returns the session reference, which holds the session ID and the
	// revision number represented by this instance.
	Ref() SessionRef

	// Refresh returns the latest revision of the session.
	//
	// If IsNotFound(err) returns true, the session has been closed and rev is
	// invalid.
	Refresh(ctx context.Context) (rev Revision, err error)

	// Get returns the attribute with key k from the attribute table.
	//
	// The returned attribute is guaranteed to be correct as of Ref().Rev.
	// Non-existant attributes are equivalent to empty attributes, therefore it
	// is not an error to request a key that has never been created.
	//
	// Peers do not always have a copy of the complete attribute table. If the
	// attribute value is unknown it is fetched from the owning peer.
	//
	// If the attribute can not be retreived because it has already been
	// modified, ShouldRetry(err) returns true. To fetch the attribute value at
	// the later revision, first call Refresh() then retry the Get() on the
	// newer revision.
	//
	// If IsNotFound(err) returns true, the session has been closed and the
	// revision can not be queried.
	Get(ctx context.Context, k string) (attr Attr, err error)

	// GetMany the attributes with keys in k from the attribute table.
	//
	// The returned attributes are guaranteed to be correct as of Ref().Rev.
	// Non-existant attributes are equivalent to empty attributes, therefore it
	// is not an error to request keys that have never been created.
	//
	// Peers do not always have a copy of the complete attribute table. If any
	// of the attribute values are unknown they are fetched from the owning peer.
	//
	// If any of the attributes can not be retreived because they hav already
	// been modified, ShouldRetry(err) returns true. To fetch the attribute
	// values at the later revision, first call Refresh() then retry the
	// GetMany() on the newer revision.
	//
	// If IsNotFound(err) returns true, the session has been closed and the
	// revision can not be queried.
	//
	// If err is nil, t contains all of the attributes specified in k.
	GetMany(ctx context.Context, k ...string) (t AttrTable, err error)

	// Update atomically modifies the attribute table.
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
	Update(ctx context.Context, attrs ...Attr) (rev Revision, err error)

	// Close destroys the session.
	//
	// The session revision represented by this instance must be the latest
	// revision. If Ref().Rev is not the latest revision the closure fails;
	// ShouldRetry(err) returns true.
	Close(ctx context.Context) (err error)
}
