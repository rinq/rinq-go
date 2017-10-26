package remotesession

import (
	"context"
	"sync"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/internal/attributes"
	revisionpkg "github.com/rinq/rinq-go/src/internal/revision"
	"github.com/rinq/rinq-go/src/internal/x/syncx"
)

type session struct {
	id     ident.SessionID
	client *client

	mutex      sync.RWMutex
	highestRev ident.Revision
	cache      attrTableCache
	isClosed   bool
}

func newSession(id ident.SessionID, client *client) *session {
	return &session{
		id:     id,
		client: client,

		cache: attrTableCache{},
	}
}

func (s *session) Head(ctx context.Context) (rinq.Revision, error) {
	unlock := syncx.RLock(&s.mutex)
	defer unlock()

	if s.isClosed {
		return nil, rinq.NotFoundError{ID: s.id}
	}

	unlock()

	rev, _, err := s.client.Fetch(ctx, s.id, "", nil)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.updateState(rev, err)

	if err != nil {
		return nil, err
	}

	return &revision{
		s.id.At(s.highestRev),
		s,
	}, nil
}

func (s *session) At(rev ident.Revision) rinq.Revision {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	ref := s.id.At(rev)

	if s.isClosed {
		return revisionpkg.Closed(ref)
	}

	s.updateState(rev, nil)

	return &revision{ref, s}
}

func (s *session) Fetch(
	ctx context.Context,
	rev ident.Revision,
	ns string,
	keys ...string,
) (attributes.List, error) {
	solvedAttrs, unsolvedKeys, err := s.fetchLocal(rev, ns, keys)
	if err != nil {
		return nil, err
	} else if len(unsolvedKeys) == 0 {
		return solvedAttrs, nil
	}

	fetchedRev, fetchedAttrs, err := s.client.Fetch(ctx, s.id, ns, unsolvedKeys)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.updateState(fetchedRev, err)

	if err != nil {
		return nil, err
	}

	if len(fetchedAttrs) == 0 {
		return solvedAttrs, nil
	}

	isStaleFetch := false
	cache, isExistingNamespace := s.cache[ns]

	for _, attr := range fetchedAttrs {
		entry := cache[attr.Key]

		// Update the cache entry if the fetched revision is newer.
		if fetchedRev > entry.FetchedAt {
			if cache == nil {
				cache = attrNamespaceCache{}
			}

			cache[attr.Key] = cachedAttr{attr, fetchedRev}
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

	if !isExistingNamespace && cache != nil {
		s.cache[ns] = cache
	}

	if isStaleFetch {
		return nil, rinq.StaleFetchError{Ref: s.id.At(rev)}
	}

	return solvedAttrs, nil
}

func (s *session) TryUpdate(
	ctx context.Context,
	rev ident.Revision,
	ns string,
	attrs attributes.List,
) (rinq.Revision, error) {
	unlock := syncx.RLock(&s.mutex)
	defer unlock()

	if s.isClosed {
		return nil, rinq.NotFoundError{ID: s.id}
	}

	ref := s.id.At(rev)

	if s.highestRev > rev {
		return nil, rinq.StaleUpdateError{Ref: ref}
	}

	updateAttrs := make(attributes.List, 0, len(attrs))

	cache := s.cache[ns]

	for _, attr := range attrs {
		if entry, ok := cache[attr.Key]; ok {
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

	updatedRev, returnedAttrs, err := s.client.Update(ctx, ref, ns, updateAttrs)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.updateState(updatedRev, err)

	if err != nil {
		return nil, err
	}

	cache, isExistingNamespace := s.cache[ns]

	for _, attr := range returnedAttrs {
		entry := cache[attr.Key]
		if updatedRev > entry.FetchedAt {
			if cache == nil {
				cache = attrNamespaceCache{}
			}

			cache[attr.Key] = cachedAttr{attr, updatedRev}
		}
	}

	if !isExistingNamespace && cache != nil {
		s.cache[ns] = cache
	}

	return &revision{
		s.id.At(s.highestRev),
		s,
	}, nil
}

func (s *session) TryClear(
	ctx context.Context,
	rev ident.Revision,
	ns string,
) (rinq.Revision, error) {
	unlock := syncx.RLock(&s.mutex)
	defer unlock()

	if s.isClosed {
		return nil, rinq.NotFoundError{ID: s.id}
	}

	ref := s.id.At(rev)

	if s.highestRev > rev {
		return nil, rinq.StaleUpdateError{Ref: ref}
	}

	for _, entry := range s.cache[ns] {
		if entry.Attr.IsFrozen {
			if entry.Attr.Value == "" {
				continue
			}

			return nil, rinq.FrozenAttributesError{Ref: ref}
		}
	}

	unlock()

	updatedRev, err := s.client.Clear(ctx, ref, ns)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.updateState(updatedRev, err)

	if err != nil {
		return nil, err
	}

	cache := s.cache[ns]

	for key, entry := range cache {
		if updatedRev > entry.FetchedAt {
			entry.Attr.Value = ""
			entry.FetchedAt = updatedRev
			cache[key] = entry
		}
	}

	return &revision{
		s.id.At(s.highestRev),
		s,
	}, nil
}

func (s *session) TryDestroy(
	ctx context.Context,
	rev ident.Revision,
) error {
	unlock := syncx.RLock(&s.mutex)
	defer unlock()

	if s.isClosed {
		return rinq.NotFoundError{ID: s.id}
	}

	ref := s.id.At(rev)

	if s.highestRev > rev {
		return rinq.StaleUpdateError{Ref: ref}
	}

	unlock()

	err := s.client.Destroy(ctx, ref)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.isClosed = true

	return nil
}

func (s *session) fetchLocal(
	rev ident.Revision,
	ns string,
	keys []string,
) (
	solved attributes.List,
	unsolved []string,
	err error,
) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	count := len(keys)
	solved = make(attributes.List, 0, count)
	unsolved = make([]string, 0, count)

	cache := s.cache[ns]

	for _, key := range keys {
		if entry, ok := cache[key]; ok {
			// The attribute hadn't been created at this revision, so we know it
			// had an empty value.
			if entry.Attr.CreatedAt > rev {
				continue
			}

			// The attribute has been changed since this revision, so we can't
			// even fetch if from the remote peer.
			if entry.Attr.UpdatedAt > rev {
				err = rinq.StaleFetchError{Ref: s.id.At(rev)}
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

	if len(unsolved) > 0 && s.isClosed {
		err = rinq.NotFoundError{ID: s.id}
	}

	return
}

func (s *session) updateState(rev ident.Revision, err error) {
	if err != nil {
		if rinq.IsNotFound(err) {
			s.isClosed = true
		}
	} else if rev > s.highestRev {
		s.highestRev = rev
	}
}
