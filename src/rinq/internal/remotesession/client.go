package remotesession

import (
	"context"
	"sync/atomic"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
	"github.com/rinq/rinq-go/src/rinq/internal/attrutil"
	"github.com/rinq/rinq-go/src/rinq/internal/command"
	"github.com/rinq/rinq-go/src/rinq/internal/traceutil"
)

type client struct {
	peerID  ident.PeerID
	invoker command.Invoker
	logger  rinq.Logger
	tracer  opentracing.Tracer
	seq     uint32
}

func newClient(
	peerID ident.PeerID,
	invoker command.Invoker,
	logger rinq.Logger,
	tracer opentracing.Tracer,
) *client {
	return &client{
		peerID:  peerID,
		invoker: invoker,
		logger:  logger,
		tracer:  tracer,
	}
}

func (c *client) Fetch(
	ctx context.Context,
	sessID ident.SessionID,
	ns string,
	keys []string,
) (
	ident.Revision,
	attrmeta.List,
	error,
) {
	span, ctx := traceutil.ChildOf(ctx, c.tracer, ext.SpanKindRPCClient)
	defer span.Finish()

	traceutil.SetupSessionFetch(span, ns, sessID)
	traceutil.LogSessionFetchRequest(span, keys)

	out := rinq.NewPayload(fetchRequest{
		Seq:       sessID.Seq,
		Namespace: ns,
		Keys:      keys,
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
		traceutil.LogSessionError(span, err)

		if rinq.IsFailureType(notFoundFailure, err) {
			err = rinq.NotFoundError{ID: sessID}
		}

		return 0, nil, err
	}

	var rsp fetchResponse
	err = in.Decode(&rsp)

	if err != nil {
		traceutil.LogSessionError(span, err)

		return 0, nil, err
	}

	traceutil.LogSessionFetchSuccess(span, rsp.Rev, rsp.Attrs)

	return rsp.Rev, rsp.Attrs, nil
}

func (c *client) Update(
	ctx context.Context,
	ref ident.Ref,
	ns string,
	attrs attrutil.List,
) (
	ident.Revision,
	attrmeta.List,
	error,
) {
	span, ctx := traceutil.ChildOf(ctx, c.tracer, ext.SpanKindRPCClient)
	defer span.Finish()

	traceutil.SetupSessionUpdate(span, ns, ref.ID)
	traceutil.LogSessionUpdateRequest(span, ref.Rev, attrs)

	out := rinq.NewPayload(updateRequest{
		Seq:       ref.ID.Seq,
		Rev:       ref.Rev,
		Namespace: ns,
		Attrs:     attrs,
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
		traceutil.LogSessionError(span, err)

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
		traceutil.LogSessionError(span, err)

		return 0, nil, err
	}

	diff := attrmeta.NewDiff(ns, rsp.Rev, len(attrs))

	for index, attr := range attrs {
		diff.Append(
			attrmeta.Attr{
				Attr:      attr,
				CreatedAt: rsp.CreatedRevs[index],
				UpdatedAt: rsp.Rev,
			},
		)
	}

	logUpdate(ctx, c.logger, c.peerID, ref.ID.At(rsp.Rev), diff)
	traceutil.LogSessionUpdateSuccess(span, rsp.Rev, diff)

	return rsp.Rev, diff.Attrs, nil
}

func (c *client) Close(
	ctx context.Context,
	ref ident.Ref,
) error {
	span, ctx := traceutil.ChildOf(ctx, c.tracer, ext.SpanKindRPCClient)
	defer span.Finish()

	traceutil.SetupSessionDestroy(span, ref.ID)
	traceutil.LogSessionDestroyRequest(span, ref.Rev)

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
		traceutil.LogSessionError(span, err)

		if rinq.IsFailureType(notFoundFailure, err) {
			err = rinq.NotFoundError{ID: ref.ID}
		} else if rinq.IsFailureType(staleUpdateFailure, err) {
			err = rinq.StaleUpdateError{Ref: ref}
		}

		return err
	}

	logClose(ctx, c.logger, c.peerID, ref)
	traceutil.LogSessionDestroySuccess(span)

	return nil
}

func (c *client) nextMessageID() ident.MessageID {
	seq := atomic.AddUint32(&c.seq, 1)

	return c.peerID.Session(0).At(0).Message(seq)
}
