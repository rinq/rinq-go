package overpass

import "context"

// Revision describes a session at a specific revision.
type Revision interface {
	// Ref returns the session reference, which holds the session ID and the
	// revision number represented by this instance.
	Ref() SessionRef

	// Refresh fetches the latest revision of the session.
	//
	// Session information may not be available locally, in which case it is
	// fetched from the peer that owns the session.
	Refresh(ctx context.Context) (Revision, error)

	// Get returns a value from the session's attribute table.
	//
	// The value is correct as of Ref().Rev.
	//
	// Peers do not always have a copy of the complete attribute table. If the
	// attribute value is unknown it is fetched from the peer that owns the
	// session. If the attribute has been changed since Ref().Rev, the
	// value can not be retreived; ShouldRetry(err) returns true.
	Get(ctx context.Context, key string) (Attr, error)

	// GetMany returns values from the session's attribute table.
	//
	// The values are correct as of Ref().Rev.
	//
	// Peers do not always have a copy of the complete attribute table. If any
	// of the attribute values are unknown they are fetched from the peer that
	// owns the session. If any of the attributes have been changed since
	// Ref().Rev, the values can not be retreived; ShouldRetry(err) returns true.
	GetMany(ctx context.Context, keys ...string) (AttrTable, error)

	// Update applies a set of changes to a session.
	//
	// Changes are atomic; either all changes in the set are applied, or the
	// session is unchanged. On success, a new session revision is produced and
	// returned.
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
	// An empty change-set produces no update and no error, the current session
	// instance is returned unchanged.
	//
	// As a convenience, if the update fails for any reason, the returned
	// Revision is the this revision. This allows the caller to assign the
	// return value over the top of an existing variable without first checking
	// for errors.
	Update(ctx context.Context, attrs ...Attr) (Revision, error)

	// Close terminates the session.
	//
	// The session revision represented by this instance must be the latest
	// revision. If Ref().Rev is not the latest revision the closure fails;
	// ShouldRetry(err) returns true.
	Close(ctx context.Context) error
}
