package localsession

import (
	"context"

	"github.com/jmalloc/twelf/src/twelf"
	"github.com/rinq/rinq-go/src/internal/attributes"
	"github.com/rinq/rinq-go/src/internal/namespaces"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/trace"
)

type revision struct {
	ref     ident.Ref
	session *Session
	attrs   attributes.Catalog
	logger  twelf.Logger
}

func (r *revision) SessionID() ident.SessionID {
	return r.ref.ID
}

func (r *revision) Refresh(ctx context.Context) (rinq.Revision, error) {
	return r.session.CurrentRevision(), nil
}

func (r *revision) Get(ctx context.Context, ns, key string) (rinq.Attr, error) {
	namespaces.MustValidate(ns)

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
	namespaces.MustValidate(ns)

	attrs := r.attrs[ns]
	table := attributes.Table{}

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
	namespaces.MustValidate(ns)

	if len(attrs) == 0 {
		return r, nil
	}

	rev, diff, err := r.session.TryUpdate(r.ref.Rev, ns, attrs)
	if err != nil {
		return r, err
	}

	logUpdate(ctx, r.logger, r.ref.ID.At(diff.Revision), diff)

	return rev, nil
}

func (r *revision) Clear(ctx context.Context, ns string) (rinq.Revision, error) {
	namespaces.MustValidate(ns)

	rev, diff, err := r.session.TryClear(r.ref.Rev, ns)
	if err != nil {
		return r, err
	}

	logClear(ctx, r.logger, r.ref.ID.At(diff.Revision), diff)

	return rev, nil
}

func (r *revision) Destroy(ctx context.Context) error {
	first, err := r.session.TryDestroy(r.ref.Rev)
	if err != nil {
		return err
	}

	if first {
		logSessionDestroy(r.logger, r.ref, r.attrs, trace.Get(ctx))
	}

	return nil
}
