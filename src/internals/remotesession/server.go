package remotesession

import (
	"context"
	"errors"

	"github.com/over-pass/overpass-go/src/internals/attrmeta"
	"github.com/over-pass/overpass-go/src/internals/command"
	"github.com/over-pass/overpass-go/src/internals/localsession"
	"github.com/over-pass/overpass-go/src/overpass"
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
	cmd overpass.Command,
	res overpass.Responder,
) {
	defer cmd.Payload.Close()

	switch cmd.Command {
	case fetchCommand:
		s.fetch(ctx, cmd, res)
	default:
		res.Error(errors.New("unknown command"))
	}
}

func (s *server) fetch(
	ctx context.Context,
	cmd overpass.Command,
	res overpass.Responder,
) {
	var req fetchRequest

	if err := cmd.Payload.Decode(&req); err != nil {
		res.Error(err)
		return
	}

	sessID := overpass.SessionID{Peer: s.peerID, Seq: req.Seq}
	_, cat, ok := s.sessions.Get(sessID)
	if !ok {
		res.Fail("not-found", "")
		return
	}

	ref, attrs := cat.Attrs()
	rsp := fetchResponse{Rev: ref.Rev}
	count := len(req.Keys)

	if count != 0 {
		rsp.Attrs = make([]attrmeta.Attr, 0, count)
		for _, key := range req.Keys {
			if attr, ok := attrs[key]; ok {
				rsp.Attrs = append(rsp.Attrs, attr)
			}
		}
	}

	payload := overpass.NewPayload(rsp)
	defer payload.Close()

	res.Done(payload)
}