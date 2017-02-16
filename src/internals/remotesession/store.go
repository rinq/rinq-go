package remotesession

import (
	"github.com/over-pass/overpass-go/src/internals/command"
	revisionpkg "github.com/over-pass/overpass-go/src/internals/revision"
	"github.com/over-pass/overpass-go/src/overpass"
)

type store struct {
	client *client
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
	}
}

func (s *store) GetRevision(ref overpass.SessionRef) (overpass.Revision, error) {
	return &revision{
		ref:    ref,
		client: s.client,
	}, nil
}
