package revisions

import (
	"context"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// Closed returns an revision that behaves as though its session has been closed.
func Closed(id ident.SessionID) rinq.Revision {
	return closed(id)
}

type closed ident.SessionID

func (r closed) SessionID() ident.SessionID {
	return ident.SessionID(r)
}

func (r closed) Refresh(context.Context) (rinq.Revision, error) {
	return nil, rinq.NotFoundError{ID: ident.SessionID(r)}
}

func (r closed) Get(context.Context, string, string) (rinq.Attr, error) {
	return rinq.Attr{}, rinq.NotFoundError{ID: ident.SessionID(r)}
}

func (r closed) GetMany(context.Context, string, ...string) (rinq.AttrTable, error) {
	return nil, rinq.NotFoundError{ID: ident.SessionID(r)}
}

func (r closed) Update(context.Context, string, ...rinq.Attr) (rinq.Revision, error) {
	return r, rinq.NotFoundError{ID: ident.SessionID(r)}
}

func (r closed) Clear(context.Context, string) (rinq.Revision, error) {
	return r, rinq.NotFoundError{ID: ident.SessionID(r)}
}

func (r closed) Destroy(context.Context) error {
	return nil
}
