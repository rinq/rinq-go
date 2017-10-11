package remotesession

import (
	"context"
	"errors"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
	"github.com/rinq/rinq-go/src/rinq/internal/command"
	"github.com/rinq/rinq-go/src/rinq/internal/localsession"
	"github.com/rinq/rinq-go/src/rinq/internal/traceutil"
)

type server struct {
	peerID   ident.PeerID
	sessions localsession.Store
	logger   rinq.Logger
}

// Listen attaches a new remote session service to the given command server.
func Listen(
	svr command.Server,
	peerID ident.PeerID,
	sessions localsession.Store,
	logger rinq.Logger,
) error {
	s := &server{
		peerID:   peerID,
		sessions: sessions,
		logger:   logger,
	}

	_, err := svr.Listen(sessionNamespace, s.handle)
	return err
}

func (s *server) handle(
	ctx context.Context,
	req rinq.Request,
	res rinq.Response,
) {
	defer req.Payload.Close()

	switch req.Command {
	case fetchCommand:
		s.fetch(ctx, req, res)
	case updateCommand:
		s.update(ctx, req, res)
	case clearCommand:
		s.clear(ctx, req, res)
	case destroyCommand:
		s.destroy(ctx, req, res)
	default:
		res.Error(errors.New("unknown command"))
	}
}

func (s *server) fetch(
	ctx context.Context,
	req rinq.Request,
	res rinq.Response,
) {
	span := opentracing.SpanFromContext(ctx)

	var args fetchRequest

	if err := req.Payload.Decode(&args); err != nil {
		res.Error(err)
		traceutil.LogSessionError(span, err)
		return
	}

	sessID := s.peerID.Session(args.Seq)

	traceutil.SetupSessionFetch(span, args.Namespace, sessID)
	traceutil.LogSessionFetchRequest(span, args.Keys)

	_, cat, ok := s.sessions.Get(sessID)
	if !ok {
		err := res.Fail(notFoundFailure, "")
		traceutil.LogSessionError(span, err)
		return
	}

	ref, attrs := cat.AttrsIn(args.Namespace)
	rsp := fetchResponse{Rev: ref.Rev}
	count := len(args.Keys)

	if count != 0 {
		rsp.Attrs = make([]attrmeta.Attr, 0, count)
		for _, key := range args.Keys {
			if attr, ok := attrs[key]; ok {
				rsp.Attrs = append(rsp.Attrs, attr)
			}
		}
	}

	payload := rinq.NewPayload(rsp)
	defer payload.Close()

	res.Done(payload)

	traceutil.LogSessionFetchSuccess(span, rsp.Rev, rsp.Attrs)
}

func (s *server) update(
	ctx context.Context,
	req rinq.Request,
	res rinq.Response,
) {
	span := opentracing.SpanFromContext(ctx)

	var args updateRequest

	if err := req.Payload.Decode(&args); err != nil {
		res.Error(err)
		traceutil.LogSessionError(span, err)
		return
	}

	sessID := s.peerID.Session(args.Seq)

	traceutil.SetupSessionUpdate(span, args.Namespace, sessID)
	traceutil.LogSessionUpdateRequest(span, args.Rev, args.Attrs)

	_, cat, ok := s.sessions.Get(sessID)
	if !ok {
		err := res.Fail(notFoundFailure, "")
		traceutil.LogSessionError(span, err)
		return
	}

	rev, diff, err := cat.TryUpdate(sessID.At(args.Rev), args.Namespace, args.Attrs)
	if err != nil {
		res.Error(errorToFailure(err))
		traceutil.LogSessionError(span, err)
		return
	}

	logRemoteUpdate(ctx, s.logger, rev.Ref(), req.Source.Ref().ID.Peer, diff)

	rsp := updateResponse{
		Rev:         rev.Ref().Rev,
		CreatedRevs: make([]ident.Revision, 0, len(args.Attrs)),
	}
	_, attrs := cat.AttrsIn(args.Namespace)

	for _, attr := range args.Attrs {
		rsp.CreatedRevs = append(
			rsp.CreatedRevs,
			attrs[attr.Key].CreatedAt,
		)
	}

	payload := rinq.NewPayload(rsp)
	defer payload.Close()

	res.Done(payload)

	traceutil.LogSessionUpdateSuccess(span, rsp.Rev, diff)
}

func (s *server) clear(
	ctx context.Context,
	req rinq.Request,
	res rinq.Response,
) {
	span := opentracing.SpanFromContext(ctx)

	var args updateRequest

	if err := req.Payload.Decode(&args); err != nil {
		res.Error(err)
		traceutil.LogSessionError(span, err)
		return
	}

	sessID := s.peerID.Session(args.Seq)

	traceutil.SetupSessionClear(span, args.Namespace, sessID)
	traceutil.LogSessionClearRequest(span, args.Rev)

	_, cat, ok := s.sessions.Get(sessID)
	if !ok {
		err := res.Fail(notFoundFailure, "")
		traceutil.LogSessionError(span, err)
		return
	}

	rev, diff, err := cat.TryClear(sessID.At(args.Rev), args.Namespace)
	if err != nil {
		res.Error(errorToFailure(err))
		traceutil.LogSessionError(span, err)
		return
	}

	logRemoteClear(ctx, s.logger, rev.Ref(), req.Source.Ref().ID.Peer, diff)

	rsp := updateResponse{
		Rev: rev.Ref().Rev,
	}

	payload := rinq.NewPayload(rsp)
	defer payload.Close()

	res.Done(payload)

	traceutil.LogSessionClearSuccess(span, rsp.Rev, diff)
}

func (s *server) destroy(
	ctx context.Context,
	req rinq.Request,
	res rinq.Response,
) {
	span := opentracing.SpanFromContext(ctx)

	var args destroyRequest

	if err := req.Payload.Decode(&args); err != nil {
		res.Error(err)
		traceutil.LogSessionError(span, err)
		return
	}

	sessID := s.peerID.Session(args.Seq)

	traceutil.SetupSessionDestroy(span, sessID)
	traceutil.LogSessionDestroyRequest(span, args.Rev)

	_, cat, ok := s.sessions.Get(sessID)
	if !ok {
		err := res.Fail(notFoundFailure, "")
		traceutil.LogSessionError(span, err)
		return
	}

	ref := sessID.At(args.Rev)

	if err := cat.TryDestroy(ref); err != nil {
		res.Error(errorToFailure(err))
		traceutil.LogSessionError(span, err)
		return
	}

	logRemoteDestroy(ctx, s.logger, cat, req.Source.Ref().ID.Peer)

	res.Close()

	traceutil.LogSessionDestroySuccess(span)
}
