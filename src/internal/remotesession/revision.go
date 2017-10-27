package remotesession

import (
	"context"

	"github.com/rinq/rinq-go/src/internal/attributes"
	"github.com/rinq/rinq-go/src/internal/namespaces"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

type revision struct {
	ref     ident.Ref
	session *session
}

func (r *revision) Ref() ident.Ref {
	return r.ref
}

func (r *revision) Refresh(ctx context.Context) (rinq.Revision, error) {
	return r.session.Head(ctx)
}

func (r *revision) Get(ctx context.Context, ns, key string) (rinq.Attr, error) {
	namespaces.MustValidate(ns)

	if r.ref.Rev == 0 {
		return rinq.Attr{Key: key}, nil
	}

	attrs, err := r.session.Fetch(ctx, r.ref.Rev, ns, key)
	if err != nil {
		return rinq.Attr{}, err
	} else if len(attrs) == 0 {
		return rinq.Attr{Key: key}, nil
	}

	return attrs[0], nil
}

func (r *revision) GetMany(ctx context.Context, ns string, keys ...string) (rinq.AttrTable, error) {
	namespaces.MustValidate(ns)

	table := attributes.Table{}

	for _, key := range keys {
		table[key] = rinq.Attr{Key: key}
	}

	if r.ref.Rev == 0 {
		return table, nil
	}

	attrs, err := r.session.Fetch(ctx, r.ref.Rev, ns, keys...)
	if err != nil {
		return nil, err
	}

	for _, attr := range attrs {
		table[attr.Key] = attr
	}

	return table, nil
}

func (r *revision) Update(ctx context.Context, ns string, attrs ...rinq.Attr) (rinq.Revision, error) {
	namespaces.MustValidate(ns)

	rev, err := r.session.TryUpdate(ctx, r.ref.Rev, ns, attrs)
	if err != nil {
		return r, err
	}

	return rev, nil
}

func (r *revision) Clear(ctx context.Context, ns string) (rinq.Revision, error) {
	namespaces.MustValidate(ns)

	rev, err := r.session.TryClear(ctx, r.ref.Rev, ns)
	if err != nil {
		return r, err
	}

	return rev, nil
}

func (r *revision) Destroy(ctx context.Context) error {
	return r.session.TryDestroy(ctx, r.ref.Rev)
}
