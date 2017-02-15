package overpassamqp

import (
	"context"
	"fmt"

	"github.com/over-pass/overpass-go/src/internals"
	"github.com/over-pass/overpass-go/src/overpass"
)

type localRevision struct {
	session *localSession
	ref     overpass.SessionRef
	attrs   internals.AttrTableWithMetaData
	diff    string
}

func (r *localRevision) Ref() overpass.SessionRef {
	return r.ref
}

func (r *localRevision) Refresh(ctx context.Context) (overpass.Revision, error) {
	return r.session.CurrentRevision()
}

func (r *localRevision) Get(ctx context.Context, key string) (overpass.Attr, error) {
	if entry, ok := r.attrs[key]; ok {
		return entry.Attr, nil
	}

	return overpass.Attr{Key: key}, nil
}

func (r *localRevision) GetMany(ctx context.Context, keys ...string) (overpass.AttrTable, error) {
	attrs := overpass.AttrTable{}

	for _, key := range keys {
		if entry, ok := r.attrs[key]; ok && entry.Attr.Value != "" {
			attrs[key] = entry.Attr
		}
	}

	return attrs, nil
}

func (r *localRevision) Update(ctx context.Context, attrs ...overpass.Attr) (overpass.Revision, error) {
	if len(attrs) == 0 {
		return r, nil
	}

	next := &localRevision{
		session: r.session,
		ref:     r.ref,
		attrs:   internals.AttrTableWithMetaData{}, // TODO implement clone function
	}
	next.ref.Rev++

	for key, entry := range r.attrs {
		next.attrs[key] = entry
	}

	var frozen []string
	for _, attr := range attrs {
		entry, ok := next.attrs[attr.Key]
		if ok {
			if attr.Value == entry.Attr.Value {
				continue
			} else if entry.Attr.IsFrozen {
				frozen = append(frozen, attr.Key)
				continue
			}
		} else {
			entry.CreatedAt = next.ref.Rev
		}

		entry.UpdatedAt = next.ref.Rev
		entry.Attr = attr
		next.attrs[attr.Key] = entry

		if next.diff != "" {
			next.diff += ", "
		}

		if attr.Value == "" {
			if attr.IsFrozen {
				next.diff += "!" + attr.Key
			} else {
				next.diff += "-" + attr.Key
			}
		} else {
			if entry.CreatedAt == entry.UpdatedAt {
				next.diff += "+"
			}

			if attr.IsFrozen {
				next.diff += attr.Key + "@" + attr.Value
			} else {
				next.diff += attr.Key + "=" + attr.Value
			}
		}
	}

	if len(frozen) > 0 {
		return r, overpass.FrozenAttributesError{Ref: r.ref, Keys: frozen}
	}

	if err := r.session.ApplyUpdate(ctx, next); err != nil {
		return r, err
	}

	return next, nil
}

func (r *localRevision) Close(ctx context.Context) error {
	return r.session.ApplyClose(ctx, r.ref)
}

func (r *localRevision) At(ref overpass.SessionRef) (overpass.Revision, error) {
	if ref.After(r.ref) {
		return nil, fmt.Errorf("%s is in the future", ref)
	} else if ref.Before(r.ref) {
		return internals.NewHistoricalRevision(r.session, ref, r.attrs), nil
	}

	return r, nil
}

func (r *localRevision) Match(constraint overpass.Constraint) bool {
	if len(constraint) == 0 {
		return true
	}

	for key, value := range constraint {
		if r.attrs[key].Attr.Value != value {
			return false
		}
	}

	return true
}

func (r *localRevision) String() string {
	return r.diff
}
