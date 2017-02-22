package amqputil

import (
	"github.com/over-pass/overpass-go/src/internals/service"
	"github.com/streadway/amqp"
)

// brokerService exposes an AMQP connection as a service.
type brokerService struct {
	service.Service
	closer *service.Closer

	broker *amqp.Connection
}

// NewBrokerService returns a service.Service based on an AMQP connection.
func NewBrokerService(broker *amqp.Connection) service.Service {
	svc, closer := service.NewImpl()

	s := &brokerService{
		Service: svc,
		closer:  closer,

		broker: broker,
	}

	go s.monitor()

	return s
}

func (s *brokerService) monitor() {
	closed := s.broker.NotifyClose(make(chan *amqp.Error))

	select {
	case err := <-closed:
		s.closer.Close(err)
		// TODO: log
	case <-s.closer.Stop():
		s.broker.Close()
		s.closer.Close(nil)
		// TODO: log
	}
}
