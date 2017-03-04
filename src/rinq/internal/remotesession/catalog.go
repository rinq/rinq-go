package remotesession

import (
	"context"
	"sync"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
	revisionpkg "github.com/rinq/rinq-go/src/rinq/internal/revision"
	"github.com/rinq/rinq-go/src/rinq/internal/syncutil"
)

type catalog struct {
	id     ident.SessionID
	client *client

	mutex      sync.RWMutex
	highestRev ident.Revision
	cache      map[string]attrCacheEntry
	isClosed   bool
}

func newCatalog(id ident.SessionID, client *client) *catalog {
	return &catalog{
		id:     id,
		client: client,

		cache: map[string]attrCacheEntry{},
	}
}

type attrCacheEntry struct {
	Attr      attrmeta.Attr
	FetchedAt ident.Revision
}

func (c *catalog) Head(ctx context.Context) (rinq.Revision, error) {
	unlock := syncutil.RLock(&c.mutex)
	defer unlock()

	if c.isClosed {
		return nil, rinq.NotFoundError{ID: c.id}
	}

	unlock()

	rev, _, err := c.client.Fetch(ctx, c.id, nil)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.updateState(rev, err)

	if err != nil {
		return nil, err
	}

	return &revision{
		ref:     c.id.At(c.highestRev),
		catalog: c,
	}, nil
}

func (c *catalog) At(rev ident.Revision) rinq.Revision {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ref := c.id.At(rev)

	if c.isClosed {
		return revisionpkg.Closed(ref)
	}

	c.updateState(rev, nil)

	return &revision{
		ref:     ref,
		catalog: c,
	}
}

func (c *catalog) Fetch(
	ctx context.Context,
	rev ident.Revision,
	keys ...string,
) ([]rinq.Attr, error) {
	solvedAttrs, unsolvedKeys, err := c.fetchLocal(rev, keys)
	if err != nil {
		return nil, err
	} else if len(unsolvedKeys) == 0 {
		return solvedAttrs, nil
	}

	fetchedRev, fetchedAttrs, err := c.client.Fetch(ctx, c.id, unsolvedKeys)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.updateState(fetchedRev, err)

	if err != nil {
		return nil, err
	}

	if len(fetchedAttrs) == 0 {
		return solvedAttrs, nil
	}

	isStaleFetch := false

	for _, attr := range fetchedAttrs {
		// Update the cache entry if the fetched revision is newer.
		entry := c.cache[attr.Key]
		if fetchedRev > entry.FetchedAt {
			c.cache[attr.Key] = attrCacheEntry{attr, fetchedRev}
		}

		if isStaleFetch {
			continue
		}

		// The attribute hadn't been created at this revision, so we know it
		// had an empty value.
		if attr.CreatedAt > rev {
			continue
		}

		// The attribute has been changed since this revision, so we know it's
		// stale, but we continue through the loop to cache any other attributes.
		if attr.UpdatedAt > rev {
			isStaleFetch = true
			continue
		}

		// We just fetch the attribute, so we know it's valid right now.
		solvedAttrs = append(solvedAttrs, attr.Attr)
	}

	if isStaleFetch {
		return nil, rinq.StaleFetchError{Ref: c.id.At(rev)}
	}

	return solvedAttrs, nil
}

func (c *catalog) TryUpdate(
	ctx context.Context,
	rev ident.Revision,
	attrs []rinq.Attr,
) (rinq.Revision, error) {
	unlock := syncutil.RLock(&c.mutex)
	defer unlock()

	if c.isClosed {
		return nil, rinq.NotFoundError{ID: c.id}
	}

	ref := c.id.At(rev)

	if c.highestRev > rev {
		return nil, rinq.StaleUpdateError{Ref: ref}
	}

	updateAttrs := make([]rinq.Attr, 0, len(attrs))

	for _, attr := range attrs {
		if entry, ok := c.cache[attr.Key]; ok {
			if entry.Attr.IsFrozen {
				if attr == entry.Attr.Attr {
					continue
				}

				return nil, rinq.FrozenAttributesError{Ref: ref}
			}

			if entry.FetchedAt == rev && attr == entry.Attr.Attr {
				continue
			}
		}

		updateAttrs = append(updateAttrs, attr)
	}

	unlock()

	updatedRev, returnedAttrs, err := c.client.Update(ctx, ref, updateAttrs)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.updateState(updatedRev, err)

	if err != nil {
		return nil, err
	}

	for _, attr := range returnedAttrs {
		entry := c.cache[attr.Key]
		if updatedRev > entry.FetchedAt {
			c.cache[attr.Key] = attrCacheEntry{attr, updatedRev}
		}
	}

	return &revision{
		ref:     c.id.At(c.highestRev),
		catalog: c,
	}, nil
}

func (c *catalog) TryDestroy(
	ctx context.Context,
	rev ident.Revision,
) error {
	unlock := syncutil.RLock(&c.mutex)
	defer unlock()

	if c.isClosed {
		return rinq.NotFoundError{ID: c.id}
	}

	ref := c.id.At(rev)

	if c.highestRev > rev {
		return rinq.StaleUpdateError{Ref: ref}
	}

	unlock()

	err := c.client.Close(ctx, ref)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.isClosed = true

	return nil
}

func (c *catalog) fetchLocal(
	rev ident.Revision,
	keys []string,
) (
	solved []rinq.Attr,
	unsolved []string,
	err error,
) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	count := len(keys)
	solved = make([]rinq.Attr, 0, count)
	unsolved = make([]string, 0, count)

	for _, key := range keys {
		if entry, ok := c.cache[key]; ok {
			// The attribute hadn't been created at this revision, so we know it
			// had an empty value.
			if entry.Attr.CreatedAt > rev {
				continue
			}

			// The attribute has been changed since this revision, so we can't
			// even fetch if from the remote peer.
			if entry.Attr.UpdatedAt > rev {
				err = rinq.StaleFetchError{Ref: c.id.At(rev)}
				return
			}

			// The attribute has been frozen, so it can't have changed, or we
			// already know the cache data is valid at or after the requested
			// revision.
			if entry.Attr.IsFrozen || rev <= entry.FetchedAt {
				solved = append(solved, entry.Attr.Attr)
				continue
			}
		}

		unsolved = append(unsolved, key)
	}

	if len(unsolved) > 0 && c.isClosed {
		err = rinq.NotFoundError{ID: c.id}
	}

	return
}

func (c *catalog) updateState(rev ident.Revision, err error) {
	if err != nil {
		if rinq.IsNotFound(err) {
			c.isClosed = true
		}
	} else if rev > c.highestRev {
		c.highestRev = rev
	}
}
