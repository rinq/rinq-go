package internals

import (
	"context"

	"github.com/over-pass/overpass-go/src/overpass"
)

// NewClosedRevision returns an revision that behaves as though its session has
// been closed.
func NewClosedRevision(ref overpass.SessionRef) overpass.Revision {
	return closedRevision(ref)
}

type closedRevision overpass.SessionRef

func (r closedRevision) Ref() overpass.SessionRef {
	return overpass.SessionRef(r)
}

func (r closedRevision) Refresh(context.Context) (overpass.Revision, error) {
	return nil, overpass.NotFoundError{ID: r.ID}
}

func (r closedRevision) Get(context.Context, string) (overpass.Attr, error) {
	return overpass.Attr{}, overpass.NotFoundError{ID: r.ID}
}

func (r closedRevision) GetMany(context.Context, ...string) (overpass.AttrTable, error) {
	return nil, overpass.NotFoundError{ID: r.ID}
}

func (r closedRevision) Update(context.Context, ...overpass.Attr) (overpass.Revision, error) {
	return r, overpass.NotFoundError{ID: r.ID}
}

func (r closedRevision) Close(context.Context) error {
	return nil
}
