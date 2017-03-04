package localsession

import (
	"context"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
	"github.com/rinq/rinq-go/src/rinq/internal/bufferpool"
)

type revision struct {
	ref     ident.Ref
	catalog Catalog
	attrs   attrmeta.Table
	logger  rinq.Logger
}

func (r *revision) Ref() ident.Ref {
	return r.ref
}

func (r *revision) Refresh(ctx context.Context) (rinq.Revision, error) {
	return r.catalog.Head(), nil
}

func (r *revision) Get(ctx context.Context, key string) (rinq.Attr, error) {
	if r.ref.Rev == 0 {
		return rinq.Attr{Key: key}, nil
	}

	attr, ok := r.attrs[key]

	// The attribute hadn't yet been created at this revision.
	if !ok || attr.CreatedAt > r.ref.Rev {
		return rinq.Attr{Key: key}, nil
	}

	// The attribute exists, but has been updated since this revision.
	// The value at this revision is no longer known.
	if attr.UpdatedAt > r.ref.Rev {
		return rinq.Attr{}, rinq.StaleFetchError{Ref: r.ref}
	}

	return attr.Attr, nil
}

func (r *revision) GetMany(ctx context.Context, keys ...string) (rinq.AttrTable, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	table := rinq.AttrTable{}

	for _, key := range keys {
		attr, ok := r.attrs[key]

		if !ok || attr.CreatedAt > r.ref.Rev {
			// The attribute hadn't yet been created at this revision.
			table[key] = rinq.Attr{Key: key}
		} else if attr.UpdatedAt <= r.ref.Rev {
			// The attribute was updated before this revision, it's still valid.
			table[key] = attr.Attr
		} else {
			return nil, rinq.StaleFetchError{Ref: r.ref}
		}
	}

	return table, nil
}

func (r *revision) Update(ctx context.Context, attrs ...rinq.Attr) (rinq.Revision, error) {
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

func (r *revision) Destroy(ctx context.Context) error {
	err := r.catalog.TryDestroy(r.ref)
	if err != nil {
		return err
	}

	logClose(ctx, r.logger, r.catalog)

	return nil
}
