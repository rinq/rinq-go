package revision

import "github.com/over-pass/overpass-go/src/overpass"

// Store is an interface for retreiving session revisions.
type Store interface {
	// GetRevision returns the session revision for the given ref.
	GetRevision(overpass.SessionRef) (overpass.Revision, error)
}

// NewAggregateStore returns a revision store that forwards to one of two
// other stores based on whether the requested revision is "local" or "remote".
func NewAggregateStore(peerID overpass.PeerID, local, remote Store) Store {
	return &aggregateStore{peerID, local, remote}
}

type aggregateStore struct {
	peerID overpass.PeerID
	local  Store
	remote Store
}

func (s *aggregateStore) GetRevision(ref overpass.SessionRef) (overpass.Revision, error) {
	if ref.ID.Peer == s.peerID {
		if s.local != nil {
			return s.local.GetRevision(ref)
		}
	} else if s.remote != nil {
		return s.remote.GetRevision(ref)
	}

	return Closed(ref), nil
}
