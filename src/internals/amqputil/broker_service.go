package amqputil

import (
	"sync/atomic"

	"github.com/over-pass/overpass-go/src/internals/service"
	"github.com/streadway/amqp"
)

// brokerService exposes an AMQP connection as a service.
type brokerService struct {
	broker *amqp.Connection
	done   chan struct{}
	err    atomic.Value
}

// NewBrokerService returns a service.Service based on an AMQP connection.
func NewBrokerService(broker *amqp.Connection) service.Service {
	s := &brokerService{
		broker: broker,
		done:   make(chan struct{}),
	}

	go s.monitor()

	return s
}

func (s *brokerService) Done() <-chan struct{} {
	return s.done
}

func (s *brokerService) Error() error {
	err, _ := s.err.Load().(error)
	return err
}

func (s *brokerService) Stop() {
	select {
	case <-s.done:
	default:
		s.broker.Close()
		<-s.done
	}
}

func (s *brokerService) monitor() {
	if err := <-s.broker.NotifyClose(make(chan *amqp.Error)); err != nil {
		s.err.Store(err)
	}

	close(s.done)
}
