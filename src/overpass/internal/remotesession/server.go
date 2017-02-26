package remotesession

import (
	"context"
	"errors"

	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpass/internal/attrmeta"
	"github.com/over-pass/overpass-go/src/overpass/internal/command"
	"github.com/over-pass/overpass-go/src/overpass/internal/localsession"
)

type server struct {
	peerID   overpass.PeerID
	sessions localsession.Store
}

// Listen attaches a new remote session service to the given command server.
func Listen(
	peerID overpass.PeerID,
	sessions localsession.Store,
	svr command.Server,
) error {
	s := &server{
		peerID:   peerID,
		sessions: sessions,
	}

	_, err := svr.Listen(sessionNamespace, s.handle)
	return err
}

func (s *server) handle(
	ctx context.Context,
	req overpass.Request,
	res overpass.Response,
) {
	defer req.Payload.Close()

	switch req.Command {
	case fetchCommand:
		s.fetch(ctx, req, res)
	case updateCommand:
		s.update(ctx, req, res)
	case closeCommand:
		s.close(ctx, req, res)
	default:
		res.Error(errors.New("unknown command"))
	}
}

func (s *server) fetch(
	ctx context.Context,
	req overpass.Request,
	res overpass.Response,
) {
	var args fetchRequest

	if err := req.Payload.Decode(&args); err != nil {
		res.Error(err)
		return
	}

	sessID := overpass.SessionID{Peer: s.peerID, Seq: args.Seq}
	_, cat, ok := s.sessions.Get(sessID)
	if !ok {
		res.Fail(notFoundFailure, "")
		return
	}

	ref, attrs := cat.Attrs()
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

	payload := overpass.NewPayload(rsp)
	defer payload.Close()

	res.Done(payload)
}

func (s *server) update(
	ctx context.Context,
	req overpass.Request,
	res overpass.Response,
) {
	var args updateRequest

	if err := req.Payload.Decode(&args); err != nil {
		res.Error(err)
		return
	}

	sessID := overpass.SessionID{Peer: s.peerID, Seq: args.Seq}
	_, cat, ok := s.sessions.Get(sessID)
	if !ok {
		res.Fail(notFoundFailure, "")
		return
	}

	rev, err := cat.TryUpdate(sessID.At(args.Rev), args.Attrs, nil)
	if err != nil {
		switch err.(type) {
		case overpass.NotFoundError:
			res.Fail(notFoundFailure, "")
		case overpass.StaleUpdateError:
			res.Fail(staleUpdateFailure, "")
		case overpass.FrozenAttributesError:
			res.Fail(frozenAttributesFailure, "")
		default:
			res.Error(err)
		}

		return
	}

	rsp := updateResponse{
		Rev:         rev.Ref().Rev,
		CreatedRevs: make([]overpass.RevisionNumber, 0, len(args.Attrs)),
	}
	_, attrs := cat.Attrs()

	for _, attr := range args.Attrs {
		rsp.CreatedRevs = append(
			rsp.CreatedRevs,
			attrs[attr.Key].CreatedAt,
		)
	}

	payload := overpass.NewPayload(rsp)
	defer payload.Close()
	res.Done(payload)
}

func (s *server) close(
	ctx context.Context,
	req overpass.Request,
	res overpass.Response,
) {
	var args closeRequest

	if err := req.Payload.Decode(&args); err != nil {
		res.Error(err)
		return
	}

	sessID := overpass.SessionID{Peer: s.peerID, Seq: args.Seq}
	_, cat, ok := s.sessions.Get(sessID)
	if !ok {
		res.Fail(notFoundFailure, "")
		return
	}

	if err := cat.TryClose(sessID.At(args.Rev)); err != nil {
		switch err.(type) {
		case overpass.NotFoundError:
			res.Fail(notFoundFailure, "")
		case overpass.StaleUpdateError:
			res.Fail(staleUpdateFailure, "")
		default:
			res.Error(err)
		}

		return
	}

	res.Close()
}
