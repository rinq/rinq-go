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

func (r *aggregateStore) GetRevision(ref overpass.SessionRef) (overpass.Revision, error) {
	if ref.ID.Peer == r.peerID {
		if r.local != nil {
			return r.local.GetRevision(ref)
		}
	} else if r.remote != nil {
		return r.remote.GetRevision(ref)
	}

	return Closed(ref), nil
}
