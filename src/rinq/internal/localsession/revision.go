package localsession

import (
	"context"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
)

type revision struct {
	ref     ident.Ref
	catalog Catalog
	attrs   attrmeta.Table
	logger  rinq.Logger
}

func newRevision(
	r ident.Ref,
	c Catalog,
	a attrmeta.Table,
	l rinq.Logger,
) rinq.Revision {
	return &revision{
		ref:     r,
		catalog: c,
		attrs:   a,
		logger:  l,
	}
}

func (r *revision) Ref() ident.Ref {
	return r.ref
}

func (r *revision) Refresh(ctx context.Context) (rinq.Revision, error) {
	return r.catalog.Head(), nil
}

func (r *revision) Get(ctx context.Context, ns, key string) (rinq.Attr, error) {
	if err := rinq.ValidateNamespace(ns); err != nil {
		return rinq.Attr{}, err
	}

	if r.ref.Rev == 0 {
		return rinq.Attr{Key: key}, nil
	}

	attr, ok := r.attrs[ns][key]

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

func (r *revision) GetMany(ctx context.Context, ns string, keys ...string) (rinq.AttrTable, error) {
	if err := rinq.ValidateNamespace(ns); err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, nil
	}

	attrs := r.attrs[ns]
	table := rinq.AttrTable{}

	for _, key := range keys {
		attr, ok := attrs[key]

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

func (r *revision) Update(ctx context.Context, ns string, attrs ...rinq.Attr) (rinq.Revision, error) {
	if err := rinq.ValidateNamespace(ns); err != nil {
		return nil, err
	}

	if len(attrs) == 0 {
		return r, nil
	}

	rev, diff, err := r.catalog.TryUpdate(r.ref, ns, attrs)
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

	logDestroy(ctx, r.logger, r.catalog)

	return nil
}
