package remotesession

import (
	"sync"

	"github.com/over-pass/overpass-go/src/internals/command"
	revisionpkg "github.com/over-pass/overpass-go/src/internals/revision"
	"github.com/over-pass/overpass-go/src/overpass"
)

type store struct {
	client *client

	mutex    sync.Mutex
	catalogs map[overpass.SessionID]*catalog
}

// NewStore returns a new store for revisions of remote sessions.
func NewStore(
	peerID overpass.PeerID,
	invoker command.Invoker,
) revisionpkg.Store {
	return &store{
		client: &client{
			peerID:  peerID,
			invoker: invoker,
		},
		catalogs: map[overpass.SessionID]*catalog{},
	}
}

func (s *store) GetRevision(ref overpass.SessionRef) (overpass.Revision, error) {
	cat := s.getCatalog(ref.ID)
	return cat.At(ref.Rev), nil
}

func (s *store) getCatalog(id overpass.SessionID) *catalog {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if cat, ok := s.catalogs[id]; ok {
		return cat
	}

	cat := &catalog{id: id, client: s.client}
	s.catalogs[id] = cat

	return cat
}
