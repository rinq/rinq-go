package remotesession

import (
	"context"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

type revision struct {
	ref     ident.Ref
	catalog *catalog
}

func newRevision(ref ident.Ref, cat *catalog) rinq.Revision {
	return &revision{
		ref:     ref,
		catalog: cat,
	}
}

func (r *revision) Ref() ident.Ref {
	return r.ref
}

func (r *revision) Refresh(ctx context.Context) (rinq.Revision, error) {
	return r.catalog.Head(ctx)
}

func (r *revision) Get(ctx context.Context, ns, key string) (rinq.Attr, error) {
	if err := rinq.ValidateNamespace(ns); err != nil {
		return rinq.Attr{}, err
	}

	if r.ref.Rev == 0 {
		return rinq.Attr{Key: key}, nil
	}

	attrs, err := r.catalog.Fetch(ctx, r.ref.Rev, ns, key)
	if err != nil {
		return rinq.Attr{}, err
	} else if len(attrs) == 0 {
		return rinq.Attr{Key: key}, nil
	}

	return attrs[0], nil
}

func (r *revision) GetMany(ctx context.Context, ns string, keys ...string) (rinq.AttrTable, error) {
	if err := rinq.ValidateNamespace(ns); err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, nil
	}

	table := rinq.AttrTable{}
	for _, key := range keys {
		table[key] = rinq.Attr{Key: key}
	}

	if r.ref.Rev == 0 {
		return table, nil
	}

	attrs, err := r.catalog.Fetch(ctx, r.ref.Rev, ns, keys...)
	if err != nil {
		return nil, err
	}

	for _, attr := range attrs {
		table[attr.Key] = attr
	}

	return table, nil
}

func (r *revision) Update(ctx context.Context, ns string, attrs ...rinq.Attr) (rinq.Revision, error) {
	if err := rinq.ValidateNamespace(ns); err != nil {
		return nil, err
	}

	rev, err := r.catalog.TryUpdate(ctx, r.ref.Rev, ns, attrs)
	if err != nil {
		return r, err
	}

	return rev, nil
}

func (r *revision) Destroy(ctx context.Context) error {
	return r.catalog.TryDestroy(ctx, r.ref.Rev)
}
