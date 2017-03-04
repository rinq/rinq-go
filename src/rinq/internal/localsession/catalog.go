package localsession

import (
	"bytes"
	"errors"
	"sync"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
)

// Catalog is an interface for manipulating an attribute table.
// There is a one-to-one relationship between sessions and catalogs.
type Catalog interface {
	// Ref returns the most recent session-ref.
	// The ref's revision increments each time a call to TryUpdate() succeeds.
	Ref() ident.Ref

	// NextMessageID generates a unique message ID from the current session-ref.
	NextMessageID() ident.MessageID

	// Head returns the most recent revision.
	// It is conceptually equivalent to catalog.At(catalog.Ref().Rev).
	Head() rinq.Revision

	// At returns a revision representing the catalog at a specific revision
	// number. The revision can not be newer than the current session-ref.
	At(ident.Revision) (rinq.Revision, error)

	// Attrs returns all attributes at the most recent revision.
	Attrs() (ident.Ref, attrmeta.Table)

	// TryUpdate adds or updates attributes in the attribute table and returns
	// the new head revision.
	//
	// The operation fails if ref is not the current session-ref, attrs includes
	// changes to frozen attributes, or the catalog is closed.
	//
	// A human-readable representation of the changes is written to diff, if it
	// is non-nil.
	TryUpdate(
		ref ident.Ref,
		attrs []rinq.Attr,
		diff *bytes.Buffer,
	) (rinq.Revision, error)

	// TryClose closes the catalog, preventing further updates.
	//
	// The operation fails if ref is not the current session-ref. It is not an
	// error to close an already-closed catalog.
	TryClose(ref ident.Ref) error

	// Close forcefully closes the catalog, preventing further updates.
	// It is not an error to close an already-closed catalog.
	Close()

	// Done returns a channel that is closed when the catalog is closed.
	Done() <-chan struct{}
}

type catalog struct {
	mutex  sync.RWMutex
	ref    ident.Ref
	attrs  attrmeta.Table
	seq    uint32
	done   chan struct{}
	logger rinq.Logger
}

// NewCatalog returns a catalog for the given session.
func NewCatalog(
	id ident.SessionID,
	logger rinq.Logger,
) Catalog {
	return &catalog{
		ref:    id.At(0),
		done:   make(chan struct{}),
		logger: logger,
	}
}

func (c *catalog) Ref() ident.Ref {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.ref
}

func (c *catalog) NextMessageID() ident.MessageID {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.seq++
	return ident.MessageID{Ref: c.ref, Seq: c.seq}
}

func (c *catalog) Head() rinq.Revision {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return &revision{
		ref:     c.ref,
		catalog: c,
		attrs:   c.attrs,
		logger:  c.logger,
	}
}

func (c *catalog) At(rev ident.Revision) (rinq.Revision, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.ref.Rev < rev {
		return nil, errors.New("revision is from the future")
	}

	return &revision{
		ref:     c.ref.ID.At(rev),
		catalog: c,
		attrs:   c.attrs,
		logger:  c.logger,
	}, nil
}

func (c *catalog) Attrs() (ident.Ref, attrmeta.Table) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.ref, c.attrs
}

func (c *catalog) TryUpdate(
	ref ident.Ref,
	attrs []rinq.Attr,
	diff *bytes.Buffer,
) (rinq.Revision, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	select {
	case <-c.done:
		return nil, rinq.NotFoundError{ID: c.ref.ID}
	default:
	}

	if ref != c.ref {
		return nil, rinq.StaleUpdateError{Ref: ref}
	}

	nextAttrs := c.attrs.Clone()
	nextRev := ref.Rev + 1

	for _, attr := range attrs {
		entry, exists := nextAttrs[attr.Key]

		if attr.Value == entry.Value && attr.IsFrozen == entry.IsFrozen {
			continue
		}

		if entry.IsFrozen {
			return nil, rinq.FrozenAttributesError{Ref: ref}
		}

		entry.Attr = attr
		entry.UpdatedAt = nextRev
		if !exists {
			entry.CreatedAt = nextRev
		}

		nextAttrs[attr.Key] = entry

		if diff != nil {
			if diff.Len() != 0 {
				diff.WriteString(", ")
			}
			attrmeta.WriteDiff(diff, entry)
		}
	}

	c.ref.Rev = nextRev
	c.attrs = nextAttrs
	c.seq = 0

	return &revision{
		ref:     c.ref,
		catalog: c,
		attrs:   c.attrs,
		logger:  c.logger,
	}, nil
}

func (c *catalog) TryClose(ref ident.Ref) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if ref != c.ref {
		return rinq.StaleUpdateError{Ref: ref}
	}

	select {
	case <-c.done:
	default:
		close(c.done)
	}

	return nil
}

func (c *catalog) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	select {
	case <-c.done:
	default:
		close(c.done)
	}
}

func (c *catalog) Done() <-chan struct{} {
	return c.done
}
