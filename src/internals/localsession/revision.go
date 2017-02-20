package localsession

import (
	"context"
	"log"

	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/over-pass/overpass-go/src/internals/attrmeta"
	"github.com/over-pass/overpass-go/src/internals/bufferpool"
	"github.com/over-pass/overpass-go/src/overpass"
)

type revision struct {
	ref     overpass.SessionRef
	catalog Catalog
	attrs   attrmeta.Table
	logger  *log.Logger
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
	attrs := overpass.AttrTable{}

	if r.ref.Rev == 0 {
		return attrs, nil
	}

	for _, key := range keys {
		attr, ok := r.attrs[key]

		// The attribute hadn't yet been created at this revision.
		if !ok || attr.CreatedAt > r.ref.Rev {
			continue
		}

		// The attribute exists, but has been updated since this revision.
		// The value at this revision is no longer known.
		if attr.UpdatedAt > r.ref.Rev {
			return nil, overpass.StaleFetchError{Ref: r.ref}
		}

		if attr.Value != "" {
			attrs[key] = attr.Attr
		}
	}

	return attrs, nil
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

	if corrID := amqputil.GetCorrelationID(ctx); corrID != "" {
		r.logger.Printf(
			"%s session updated {%s} [%s]",
			rev.Ref().ShortString(),
			diff.String(),
			corrID,
		)
	} else {
		r.logger.Printf(
			"%s session updated {%s}",
			rev.Ref().ShortString(),
			diff.String(),
		)
	}

	return rev, nil
}

func (r *revision) Close(ctx context.Context) error {
	return r.catalog.TryClose(r.ref)
}
