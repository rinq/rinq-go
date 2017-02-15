package internals

import "github.com/over-pass/overpass-go/src/overpass"

// RevisionStore is an interface for retreiving session revisions.
type RevisionStore interface {
	// GetRevision returns the session revision for the given ref.
	GetRevision(overpass.SessionRef) (overpass.Revision, error)
}

// NewAggregateRevisionStore returns a revision store that forwards to one of two
// other stores based on whether the requested revision is "local" or "remote".
func NewAggregateRevisionStore(
	peerID overpass.PeerID,
	local, remote RevisionStore,
) RevisionStore {
	return &aggregateRevisionStore{peerID, local, remote}
}

type aggregateRevisionStore struct {
	peerID overpass.PeerID
	local  RevisionStore
	remote RevisionStore
}

func (r *aggregateRevisionStore) GetRevision(ref overpass.SessionRef) (overpass.Revision, error) {
	if ref.ID.Peer == r.peerID {
		return r.local.GetRevision(ref)
	}

	return r.remote.GetRevision(ref)
}
