package internals

import (
	"context"

	"github.com/over-pass/overpass-go/src/overpass"
)

// NewHistoricalRevision returns an revision that exposes the attribute table as
// though it were from an older revision of the same session. It is used as the
// source revision in commands and notifications when the session's current
// revision is known to be later than revision specified in the request.
func NewHistoricalRevision(
	sess overpass.Session,
	ref overpass.SessionRef,
	attrs AttrTableWithMetaData,
) overpass.Revision {
	return &historicalRevision{sess, ref, attrs}
}

type historicalRevision struct {
	session overpass.Session
	ref     overpass.SessionRef
	attrs   AttrTableWithMetaData
}

func (r *historicalRevision) Ref() overpass.SessionRef {
	return r.ref
}

func (r *historicalRevision) Refresh(ctx context.Context) (overpass.Revision, error) {
	return r.session.CurrentRevision()
}

func (r *historicalRevision) Get(ctx context.Context, key string) (overpass.Attr, error) {
	entry, ok := r.attrs[key]

	// The attribute doesn't exist at all, or was created after this revision.
	if !ok || entry.CreatedAt > r.ref.Rev {
		return overpass.Attr{Key: key}, nil
	}

	// The attribute exists, but has been updated since this revision. The value
	// at this revision is no longer known.
	if entry.UpdatedAt > r.ref.Rev {
		return overpass.Attr{}, overpass.StaleFetchError{Ref: r.ref}
	}

	return entry.Attr, nil
}

func (r *historicalRevision) GetMany(ctx context.Context, keys ...string) (overpass.AttrTable, error) {
	attrs := overpass.AttrTable{}

	for _, key := range keys {
		entry, ok := r.attrs[key]

		// The attribute doesn't exist at all, or was created after this revision.
		if !ok || entry.CreatedAt > r.ref.Rev {
			continue
		}

		// The attribute exists, but has been updated since this revision. The
		// value at this revision is no longer known.
		if entry.UpdatedAt > r.ref.Rev {
			return nil, overpass.StaleFetchError{Ref: r.ref}
		}

		if entry.Attr.Value != "" && entry.Attr.IsFrozen {
			attrs[key] = entry.Attr
		}
	}

	return attrs, nil
}

func (r *historicalRevision) Update(ctx context.Context, attrs ...overpass.Attr) (overpass.Revision, error) {
	return r, overpass.StaleUpdateError{Ref: r.ref}
}

func (r *historicalRevision) Close(ctx context.Context) error {
	return overpass.StaleUpdateError{Ref: r.ref}
}
