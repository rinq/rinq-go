package overpassamqp

import (
	"log"
	"sync/atomic"

	"github.com/over-pass/overpass-go/src/internals"
	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/streadway/amqp"
)

// peer is an AMQP-based implementation of overpass.Peer.
type peer struct {
	id         overpass.PeerID
	broker     *amqp.Connection
	sessions   *localStore
	invoker    internals.Invoker
	server     internals.Server
	notifier   internals.Notifier
	listener   internals.Listener
	logger     *log.Logger
	sessionSeq uint32
}

func newPeer(
	id overpass.PeerID,
	broker *amqp.Connection,
	sessions *localStore,
	invoker internals.Invoker,
	server internals.Server,
	notifier internals.Notifier,
	listener internals.Listener,
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
	sessID := overpass.SessionID{
		Peer: p.id,
		Seq:  atomic.AddUint32(&p.sessionSeq, 1),
	}

	sess := newLocalSession(
		sessID,
		p.invoker,
		p.notifier,
		p.listener,
		p.logger,
	)

	p.sessions.Add(sess)

	go func() {
		<-sess.Done()
		p.sessions.Remove(sess)
	}()

	return sess
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
	return internals.Wait(
		p.invoker,
		p.server,
		p.listener,
	)
}

func (p *peer) Close() {
	p.broker.Close()
	p.Wait()
}
