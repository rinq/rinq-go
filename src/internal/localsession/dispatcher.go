package localsession

import (
	"sync"

	"github.com/rinq/rinq-go/src/internal/transport"
)

// Dispatcher reads notifications from a consumer and forwards them to all
// sessions.
type Dispatcher struct {
	Consumer transport.Consumer
	Sessions *Store
}

// Run consumes notifications from d.Notifications and calls
// Session.acceptNotification() for each of the target sessions.
func (d *Dispatcher) Run() {
	for n := range d.Consumer.Queue() {
		d.dispatch(n)
	}
}

// dispatch finds the targets for n and calls Session.acceptNotification(n)
// for each of the  target sessions.
func (d *Dispatcher) dispatch(n transport.InboundNotification) {
	var g sync.WaitGroup

	done := n.Done
	n.Done = g.Done

	d.Sessions.Each(func(s *Session) {
		g.Add(1)
		s.receiver.Accept(n)
	})

	go func() {
		g.Wait()
		n.Payload.Close()
		done()
	}()
}
