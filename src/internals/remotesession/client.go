package remotesession

import (
	"context"
	"sync/atomic"

	"github.com/over-pass/overpass-go/src/internals/attrmeta"
	"github.com/over-pass/overpass-go/src/internals/command"
	"github.com/over-pass/overpass-go/src/overpass"
)

type client struct {
	peerID  overpass.PeerID
	invoker command.Invoker
	seq     uint32
}

func (c *client) Fetch(
	ctx context.Context,
	sessID overpass.SessionID,
	keys []string,
) (
	overpass.RevisionNumber,
	[]attrmeta.Attr,
	error,
) {
	out := overpass.NewPayload(fetchRequest{
		Seq:  sessID.Seq,
		Keys: keys,
	})
	defer out.Close()

	in, err := c.invoker.CallUnicast(
		ctx,
		c.nextMessageID(),
		sessID.Peer,
		sessionNamespace,
		fetchCommand,
		out,
	)
	if err != nil {
		if overpass.IsFailureType(notFoundFailure, err) {
			err = overpass.NotFoundError{ID: sessID}
		}
		return 0, nil, err
	}

	var rsp fetchResponse
	err = in.Decode(&rsp)

	if err != nil {
		return 0, nil, err
	}

	return rsp.Rev, rsp.Attrs, nil
}

func (c *client) nextMessageID() overpass.MessageID {
	return overpass.MessageID{
		Session: overpass.SessionRef{
			ID: overpass.SessionID{Peer: c.peerID},
		},
		Seq: atomic.AddUint32(&c.seq, 1),
	}
}
