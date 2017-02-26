package remotesession

import (
	"context"
	"sync/atomic"

	"github.com/over-pass/overpass-go/src/internal/attrmeta"
	"github.com/over-pass/overpass-go/src/internal/command"
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

	_, in, err := c.invoker.CallUnicast(
		ctx,
		c.nextMessageID(),
		sessID.Peer,
		sessionNamespace,
		fetchCommand,
		out,
	)
	defer in.Close()

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

func (c *client) Update(
	ctx context.Context,
	ref overpass.SessionRef,
	attrs []overpass.Attr,
) (
	overpass.RevisionNumber,
	[]attrmeta.Attr,
	error,
) {
	out := overpass.NewPayload(updateRequest{
		Seq:   ref.ID.Seq,
		Rev:   ref.Rev,
		Attrs: attrs,
	})
	defer out.Close()

	_, in, err := c.invoker.CallUnicast(
		ctx,
		c.nextMessageID(),
		ref.ID.Peer,
		sessionNamespace,
		updateCommand,
		out,
	)
	defer in.Close()

	if err != nil {
		if overpass.IsFailureType(notFoundFailure, err) {
			err = overpass.NotFoundError{ID: ref.ID}
		} else if overpass.IsFailureType(staleUpdateFailure, err) {
			err = overpass.StaleUpdateError{Ref: ref}
		} else if overpass.IsFailureType(frozenAttributesFailure, err) {
			err = overpass.FrozenAttributesError{Ref: ref}
		}

		return 0, nil, err
	}

	var rsp updateResponse
	err = in.Decode(&rsp)

	if err != nil {
		return 0, nil, err
	}

	updatedAttrs := make([]attrmeta.Attr, 0, len(attrs))

	for index, attr := range attrs {
		updatedAttrs = append(
			updatedAttrs,
			attrmeta.Attr{
				Attr:      attr,
				CreatedAt: rsp.CreatedRevs[index],
				UpdatedAt: rsp.Rev,
			},
		)
	}

	return rsp.Rev, updatedAttrs, nil
}

func (c *client) Close(
	ctx context.Context,
	ref overpass.SessionRef,
) error {
	out := overpass.NewPayload(closeRequest{
		Seq: ref.ID.Seq,
		Rev: ref.Rev,
	})
	defer out.Close()

	_, in, err := c.invoker.CallUnicast(
		ctx,
		c.nextMessageID(),
		ref.ID.Peer,
		sessionNamespace,
		closeCommand,
		out,
	)
	defer in.Close()

	if err != nil {
		if overpass.IsFailureType(notFoundFailure, err) {
			err = overpass.NotFoundError{ID: ref.ID}
		} else if overpass.IsFailureType(staleUpdateFailure, err) {
			err = overpass.StaleUpdateError{Ref: ref}
		}

		return err
	}

	return nil
}

func (c *client) nextMessageID() overpass.MessageID {
	return overpass.MessageID{
		Session: overpass.SessionRef{
			ID: overpass.SessionID{Peer: c.peerID},
		},
		Seq: atomic.AddUint32(&c.seq, 1),
	}
}
