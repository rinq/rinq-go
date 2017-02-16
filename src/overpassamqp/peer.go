package overpassamqp

import (
	"log"
	"sync/atomic"

	"github.com/over-pass/overpass-go/src/internals/command"
	"github.com/over-pass/overpass-go/src/internals/localsession"
	"github.com/over-pass/overpass-go/src/internals/notify"
	"github.com/over-pass/overpass-go/src/internals/service"
	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/streadway/amqp"
)

// peer is an AMQP-based implementation of overpass.Peer.
type peer struct {
	id       overpass.PeerID
	broker   *amqp.Connection
	sessions localsession.Store
	invoker  command.Invoker
	server   command.Server
	notifier notify.Notifier
	listener notify.Listener
	logger   *log.Logger
	seq      uint32
}

func newPeer(
	id overpass.PeerID,
	broker *amqp.Connection,
	sessions localsession.Store,
	invoker command.Invoker,
	server command.Server,
	notifier notify.Notifier,
	listener notify.Listener,
	logger *log.Logger,
) *peer {
	return &peer{
		id:       id,
		broker:   broker,
		sessions: sessions,
		invoker:  invoker,
		server:   server,
		notifier: notifier,
		listener: listener,
		logger:   logger,
	}
}

func (p *peer) ID() overpass.PeerID {
	return p.id
}

func (p *peer) Session() overpass.Session {
	id := overpass.SessionID{
		Peer: p.id,
		Seq:  atomic.AddUint32(&p.seq, 1),
	}

	catalog := localsession.NewCatalog(id, p.logger)
	session := localsession.NewSession(
		id,
		catalog,
		p.invoker,
		p.notifier,
		p.listener,
		p.logger,
	)

	p.sessions.Add(session, catalog)

	go func() {
		<-session.Done()
		p.sessions.Remove(id)
	}()

	return session
}

func (p *peer) Listen(namespace string, handler overpass.CommandHandler) error {
	if err := overpass.ValidateNamespace(namespace); err != nil {
		return err
	}

	added, err := p.server.Listen(namespace, handler)

	if added {
		p.logger.Printf(
			"%s started listening for command requests in '%s' namespace",
			p.id.ShortString(),
			namespace,
		)
	}

	return err
}

func (p *peer) Unlisten(namespace string) error {
	if err := overpass.ValidateNamespace(namespace); err != nil {
		return err
	}

	removed, err := p.server.Unlisten(namespace)

	if removed {
		p.logger.Printf(
			"%s stopped listening for command requests in '%s' namespace",
			p.id.ShortString(),
			namespace,
		)
	}

	return err
}

func (p *peer) Wait() error {
	return service.Wait(
		p.invoker,
		p.server,
		p.listener,
	)
}

func (p *peer) Close() {
	p.broker.Close()
	p.Wait()
}
