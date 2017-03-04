package revision

import (
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// Store is an interface for retreiving session revisions.
type Store interface {
	// GetRevision returns the session revision for the given ref.
	GetRevision(ident.Ref) (rinq.Revision, error)
}

// AggregateStore is a revision store that forwards to one of two other stores
// based on whether the requested revision is considered "local" or "remote".
type AggregateStore struct {
	PeerID ident.PeerID
	Local  Store
	Remote Store
}

// GetRevision returns the session revision for the given ref.
func (s *AggregateStore) GetRevision(ref ident.Ref) (rinq.Revision, error) {
	if ref.ID.Peer == s.PeerID {
		if s.Local != nil {
			return s.Local.GetRevision(ref)
		}
	} else if s.Remote != nil {
		return s.Remote.GetRevision(ref)
	}

	return Closed(ref), nil
}
