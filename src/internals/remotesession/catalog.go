package remotesession

import (
	"context"
	"errors"
	"sync"

	"github.com/over-pass/overpass-go/src/internals/attrmeta"
	"github.com/over-pass/overpass-go/src/internals/deferutil"
	revisionpkg "github.com/over-pass/overpass-go/src/internals/revision"
	"github.com/over-pass/overpass-go/src/overpass"
)

type catalog struct {
	id     overpass.SessionID
	client *client

	mutex      sync.RWMutex
	highestRev overpass.RevisionNumber
	cache      map[string]cacheEntry
	isClosed   bool
}

type cacheEntry struct {
	Attr      attrmeta.Attr
	FetchedAt overpass.RevisionNumber
}

func (c *catalog) Head(ctx context.Context) (overpass.Revision, error) {
	unlock := deferutil.RLock(&c.mutex)
	defer unlock()

	if c.isClosed {
		return nil, overpass.NotFoundError{ID: c.id}
	}

	unlock()

	rev, _, err := c.client.Fetch(ctx, c.id, nil)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err != nil {
		if overpass.IsNotFound(err) {
			c.isClosed = true
		}
		return nil, err
	}

	if rev > c.highestRev {
		c.highestRev = rev
	}

	return &revision{
		ref:     c.id.At(c.highestRev),
		catalog: c,
	}, nil
}

func (c *catalog) At(rev overpass.RevisionNumber) overpass.Revision {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ref := c.id.At(rev)

	if c.isClosed {
		return revisionpkg.Closed(ref)
	}

	if rev > c.highestRev {
		c.highestRev = rev
	}

	return &revision{
		ref:     ref,
		catalog: c,
	}
}

func (c *catalog) Fetch(
	ctx context.Context,
	rev overpass.RevisionNumber,
	keys ...string,
) ([]overpass.Attr, error) {
	unlock := deferutil.RLock(&c.mutex)
	defer unlock()

	solvedAttrs, unsolvedKeys, err := c.fetchLocal(rev, keys)
	if err != nil {
		return nil, err
	}

	if len(unsolvedKeys) == 0 {
		return solvedAttrs, nil
	}

	if c.isClosed {
		return nil, overpass.NotFoundError{ID: c.id}
	}

	unlock()

	fetchedRev, fetchedAttrs, err := c.client.Fetch(ctx, c.id, unsolvedKeys)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err != nil {
		if overpass.IsNotFound(err) {
			c.isClosed = true
		}
		return nil, err
	}

	if fetchedRev < rev {
		return nil, errors.New("revision is from the future")
	}

	if fetchedRev > c.highestRev {
		c.highestRev = fetchedRev
	}

	if len(fetchedAttrs) == 0 {
		return solvedAttrs, nil
	}

	isStaleFetch := false

	if c.cache == nil {
		c.cache = map[string]cacheEntry{}
	}

	for _, attr := range fetchedAttrs {
		// Update the cache entry if the fetched revision is newer.
		entry := c.cache[attr.Key]
		if fetchedRev > entry.FetchedAt {
			c.cache[attr.Key] = cacheEntry{attr, fetchedRev}
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
		return nil, overpass.StaleFetchError{Ref: c.id.At(rev)}
	}

	return solvedAttrs, nil
}

func (c *catalog) TryUpdate(
	ctx context.Context,
	rev overpass.RevisionNumber,
	attrs []overpass.Attr,
) (overpass.Revision, error) {
	unlock := deferutil.RLock(&c.mutex)
	defer unlock()

	if c.isClosed {
		return nil, overpass.NotFoundError{ID: c.id}
	}

	ref := c.id.At(rev)

	if c.highestRev > rev {
		return nil, overpass.StaleUpdateError{Ref: ref}
	}

	updateAttrs := make([]overpass.Attr, 0, len(attrs))

	for _, attr := range attrs {
		if entry, ok := c.cache[attr.Key]; ok {
			if entry.Attr.IsFrozen {
				if attr == entry.Attr.Attr {
					continue
				}

				return nil, overpass.FrozenAttributesError{Ref: ref}
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

	if err != nil {
		if overpass.IsNotFound(err) {
			c.isClosed = true
		}
		return nil, err
	}

	if updatedRev > c.highestRev {
		c.highestRev = updatedRev
	}

	if c.cache == nil {
		c.cache = map[string]cacheEntry{}
	}

	for _, attr := range returnedAttrs {
		entry := c.cache[attr.Key]
		if updatedRev > entry.FetchedAt {
			c.cache[attr.Key] = cacheEntry{attr, updatedRev}
		}
	}

	return &revision{
		ref:     c.id.At(c.highestRev),
		catalog: c,
	}, nil
}

func (c *catalog) TryClose(
	ctx context.Context,
	rev overpass.RevisionNumber,
) error {
	unlock := deferutil.RLock(&c.mutex)
	defer unlock()

	if c.isClosed {
		return overpass.NotFoundError{ID: c.id}
	}

	ref := c.id.At(rev)

	if c.highestRev > rev {
		return overpass.StaleUpdateError{Ref: ref}
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
	rev overpass.RevisionNumber,
	keys []string,
) (
	solved []overpass.Attr,
	unsolved []string,
	err error,
) {
	count := len(keys)
	solved = make([]overpass.Attr, 0, count)
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
				err = overpass.StaleFetchError{Ref: c.id.At(rev)}
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

	return
}
