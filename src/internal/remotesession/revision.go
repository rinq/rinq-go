package remotesession

import (
	"context"

	"github.com/over-pass/overpass-go/src/overpass"
)

type revision struct {
	ref     overpass.SessionRef
	catalog *catalog
}

func (r *revision) Ref() overpass.SessionRef {
	return r.ref
}

func (r *revision) Refresh(ctx context.Context) (overpass.Revision, error) {
	return r.catalog.Head(ctx)
}

func (r *revision) Get(ctx context.Context, key string) (overpass.Attr, error) {
	if r.ref.Rev == 0 {
		return overpass.Attr{Key: key}, nil
	}

	attrs, err := r.catalog.Fetch(ctx, r.ref.Rev, key)
	if err != nil {
		return overpass.Attr{}, err
	} else if len(attrs) == 0 {
		return overpass.Attr{Key: key}, nil
	}

	return attrs[0], nil
}

func (r *revision) GetMany(ctx context.Context, keys ...string) (overpass.AttrTable, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	table := overpass.AttrTable{}
	for _, key := range keys {
		table[key] = overpass.Attr{Key: key}
	}

	if r.ref.Rev == 0 {
		return table, nil
	}

	attrs, err := r.catalog.Fetch(ctx, r.ref.Rev, keys...)
	if err != nil {
		return nil, err
	}

	for _, attr := range attrs {
		table[attr.Key] = attr
	}

	return table, nil
}

func (r *revision) Update(ctx context.Context, attrs ...overpass.Attr) (overpass.Revision, error) {
	rev, err := r.catalog.TryUpdate(ctx, r.ref.Rev, attrs)
	if err != nil {
		return r, err
	}

	return rev, nil
}

func (r *revision) Close(ctx context.Context) error {
	return r.catalog.TryClose(ctx, r.ref.Rev)
}
