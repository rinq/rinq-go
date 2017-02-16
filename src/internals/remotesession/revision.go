package remotesession

import (
	"context"
	"sync"

	"github.com/over-pass/overpass-go/src/internals/attrmeta"
	"github.com/over-pass/overpass-go/src/internals/deferutil"
	"github.com/over-pass/overpass-go/src/overpass"
)

type revision struct {
	ref    overpass.SessionRef
	mutex  sync.RWMutex
	attrs  attrmeta.Table
	client *client
}

func (r *revision) Ref() overpass.SessionRef {
	return r.ref
}

func (r *revision) Refresh(context.Context) (overpass.Revision, error) {
	return nil, overpass.NotFoundError{ID: r.ref.ID}
}

func (r *revision) Get(ctx context.Context, key string) (overpass.Attr, error) {
	if r.ref.Rev == 0 {
		return overpass.Attr{Key: key}, nil
	}

	var attr attrmeta.Attr
	var ok bool

	deferutil.RWith(&r.mutex, func() {
		attr, ok = r.attrs[key]
	})

	// The attribute isn't in the map, but it may actually exist.
	if !ok {
		_, fetched, err := r.client.Fetch(ctx, r.ref.ID, key)
		if err != nil {
			return overpass.Attr{}, err
		}
		if len(fetched) == 0 {
			return overpass.Attr{Key: key}, nil
		}
		attr = fetched[0]
	}

	// The attribute hadn't yet been created at this revision.
	if attr.CreatedAt > r.ref.Rev {
		return overpass.Attr{Key: key}, nil
	}

	// The attribute exists, but has been updated since this revision.
	// The value at this revision is no longer known.
	if attr.UpdatedAt > r.ref.Rev {
		return overpass.Attr{}, overpass.StaleFetchError{Ref: r.ref}
	}

	return attr.Attr, nil

	// return overpass.Attr{}, overpass.NotFoundError{ID: r.ref.ID}
}

func (r *revision) GetMany(ctx context.Context, keys ...string) (overpass.AttrTable, error) {
	attrs := overpass.AttrTable{}

	if r.ref.Rev == 0 {
		return attrs, nil
	}

	var missing []string

	for _, key := range keys {
		attr, ok := r.attrs[key]

		// The attribute isn't in the map, but it may actually exist.
		if !ok {
			missing = append(missing, key)
		}

		// The attribute hadn't yet been created at this revision.
		if attr.CreatedAt > r.ref.Rev {
			continue
		}

		// The attribute exists, but has been updated since this revision.
		// The value at this revision is no longer known.
		if attr.UpdatedAt > r.ref.Rev {
			return nil, overpass.StaleFetchError{Ref: r.ref}
		}

		attrs[attr.Key] = attr.Attr
	}

	if len(missing) > 0 {
		panic("not impl")
		// fetched, err := r.store.Fetch(s.ref.ID, missing)
		// if err != nil {
		// 	return nil, err
		// }
		//
		// for _, attr := range fetched {
		// 	attrs[attr.Key] = attr
		// }
	}

	return attrs, nil
}

func (r *revision) Update(context.Context, ...overpass.Attr) (overpass.Revision, error) {
	return r, overpass.NotFoundError{ID: r.ref.ID}
}

func (r *revision) Close(context.Context) error {
	return nil
}
