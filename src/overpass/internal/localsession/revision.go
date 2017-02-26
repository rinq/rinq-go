package localsession

import (
	"context"

	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpass/internal/attrmeta"
	"github.com/over-pass/overpass-go/src/overpass/internal/bufferpool"
)

type revision struct {
	ref     overpass.SessionRef
	catalog Catalog
	attrs   attrmeta.Table
	logger  overpass.Logger
}

func (r *revision) Ref() overpass.SessionRef {
	return r.ref
}

func (r *revision) Refresh(ctx context.Context) (overpass.Revision, error) {
	return r.catalog.Head(), nil
}

func (r *revision) Get(ctx context.Context, key string) (overpass.Attr, error) {
	if r.ref.Rev == 0 {
		return overpass.Attr{Key: key}, nil
	}

	attr, ok := r.attrs[key]

	// The attribute hadn't yet been created at this revision.
	if !ok || attr.CreatedAt > r.ref.Rev {
		return overpass.Attr{Key: key}, nil
	}

	// The attribute exists, but has been updated since this revision.
	// The value at this revision is no longer known.
	if attr.UpdatedAt > r.ref.Rev {
		return overpass.Attr{}, overpass.StaleFetchError{Ref: r.ref}
	}

	return attr.Attr, nil
}

func (r *revision) GetMany(ctx context.Context, keys ...string) (overpass.AttrTable, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	table := overpass.AttrTable{}

	for _, key := range keys {
		attr, ok := r.attrs[key]

		if !ok || attr.CreatedAt > r.ref.Rev {
			// The attribute hadn't yet been created at this revision.
			table[key] = overpass.Attr{Key: key}
		} else if attr.UpdatedAt <= r.ref.Rev {
			// The attribute was updated before this revision, it's still valid.
			table[key] = attr.Attr
		} else {
			return nil, overpass.StaleFetchError{Ref: r.ref}
		}
	}

	return table, nil
}

func (r *revision) Update(ctx context.Context, attrs ...overpass.Attr) (overpass.Revision, error) {
	if len(attrs) == 0 {
		return r, nil
	}

	diff := bufferpool.Get()
	defer bufferpool.Put(diff)

	rev, err := r.catalog.TryUpdate(r.ref, attrs, diff)
	if err != nil {
		return r, err
	}

	logUpdate(ctx, r.logger, rev.Ref(), diff)

	return rev, nil
}

func (r *revision) Close(ctx context.Context) error {
	return r.catalog.TryClose(r.ref)
}
