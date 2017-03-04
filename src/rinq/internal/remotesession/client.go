package remotesession

import (
	"context"
	"sync/atomic"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
	"github.com/rinq/rinq-go/src/rinq/internal/command"
)

type client struct {
	peerID  ident.PeerID
	invoker command.Invoker
	logger  rinq.Logger
	seq     uint32
}

func newClient(
	peerID ident.PeerID,
	invoker command.Invoker,
	logger rinq.Logger,
) *client {
	return &client{
		peerID:  peerID,
		invoker: invoker,
		logger:  logger,
	}
}

func (c *client) Fetch(
	ctx context.Context,
	sessID ident.SessionID,
	keys []string,
) (
	ident.Revision,
	[]attrmeta.Attr,
	error,
) {
	out := rinq.NewPayload(fetchRequest{
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
		if rinq.IsFailureType(notFoundFailure, err) {
			err = rinq.NotFoundError{ID: sessID}
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
	ref ident.Ref,
	attrs []rinq.Attr,
) (
	ident.Revision,
	[]attrmeta.Attr,
	error,
) {
	out := rinq.NewPayload(updateRequest{
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
		if rinq.IsFailureType(notFoundFailure, err) {
			err = rinq.NotFoundError{ID: ref.ID}
		} else if rinq.IsFailureType(staleUpdateFailure, err) {
			err = rinq.StaleUpdateError{Ref: ref}
		} else if rinq.IsFailureType(frozenAttributesFailure, err) {
			err = rinq.FrozenAttributesError{Ref: ref}
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

	logUpdate(ctx, c.logger, c.peerID, ref.ID.At(rsp.Rev), updatedAttrs)

	return rsp.Rev, updatedAttrs, nil
}

func (c *client) Close(
	ctx context.Context,
	ref ident.Ref,
) error {
	out := rinq.NewPayload(closeRequest{
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
		if rinq.IsFailureType(notFoundFailure, err) {
			err = rinq.NotFoundError{ID: ref.ID}
		} else if rinq.IsFailureType(staleUpdateFailure, err) {
			err = rinq.StaleUpdateError{Ref: ref}
		}

		return err
	}

	logClose(ctx, c.logger, c.peerID, ref)

	return nil
}

func (c *client) nextMessageID() ident.MessageID {
	return ident.MessageID{
		Ref: ident.Ref{
			ID: ident.SessionID{Peer: c.peerID},
		},
		Seq: atomic.AddUint32(&c.seq, 1),
	}
}
