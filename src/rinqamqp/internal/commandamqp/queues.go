package commandamqp

import (
	"sync"

	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/amqputil"
	"github.com/streadway/amqp"
)

// balancedRequestQueue returns the name of the queue used for balanced
// command requests in the given namespace.
func balancedRequestQueue(namespace string) string {
	return "cmd." + namespace
}

// requestQueue returns the name of the queue used for unicast and multicast
// command requests.
func requestQueue(id ident.PeerID) string {
	return id.ShortString() + ".req"
}

// responseQueue returns the name of the queue used for command responses.
func responseQueue(id ident.PeerID) string {
	return id.ShortString() + ".rsp"
}

// QueueSet declares AMQP resources for queuing balanced command requests.
type QueueSet struct {
	Channels amqputil.ChannelPool

	mutex  sync.Mutex
	queues map[string]string
}

// Get declares the AMQP queue used for balanced command requests in the given
// namespace and returns the queue name.
func (s *QueueSet) Get(namespace string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if queue, ok := s.queues[namespace]; ok {
		return queue, nil
	}

	queue := balancedRequestQueue(namespace)

	channel, err := s.Channels.Get()
	if err != nil {
		return "", err
	}
	defer s.Channels.Put(channel)

	if _, err := channel.QueueDeclare(
		queue,
		true,  // durable
		false, // autoDelete
		false, // exclusive,
		false, // noWait
		amqp.Table{"x-max-priority": priorityCount},
	); err != nil {
		return "", err
	}

	if err := channel.QueueBind(
		queue,
		namespace,
		balancedExchange,
		false, // noWait
		nil,   // args
	); err != nil {
		return "", err
	}

	if s.queues == nil {
		s.queues = map[string]string{}
	}
	s.queues[namespace] = queue

	return queue, nil
}

// DeleteIfUnused removes the AMQP queue for the given namespace if there are
// no pending messages or active consumers.
func (s *QueueSet) DeleteIfUnused(namespace string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	queue, ok := s.queues[namespace]
	if !ok {
		return nil
	}

	delete(s.queues, namespace)

	channel, err := s.Channels.Get()
	if err != nil {
		return err
	}
	defer s.Channels.Put(channel)

	_, err = channel.QueueDelete(
		queue,
		true, // ifUnused
		true, // ifEmpty
		true, // noWait
	)

	if err != nil {
		if amqpErr, ok := err.(*amqp.Error); ok {
			if amqpErr.Code == amqp.PreconditionFailed {
				// AMQP spec dictates that precondition-failed is returned
				// if the queue has pending messages or active consumers.
				// RabbitMQ, however, does not indicate an error here.
				return nil
			}
		}

		panic(err)
	}

	return err
}

// DeleteUnused removes AMQP queues known to s that have no pending messages
// or active consumers.
func (s *QueueSet) DeleteUnused() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	queues := s.queues
	s.queues = nil

	var channel *amqp.Channel

	for _, queue := range queues {
		if channel == nil {
			var err error
			channel, err = s.Channels.Get()
			if err != nil {
				return err
			}
			defer s.Channels.Put(channel)
		}

		_, err := channel.QueueDelete(
			queue,
			true, // ifUnused
			true, // ifEmpty
			true, // noWait
		)

		if err != nil {
			if amqpErr, ok := err.(*amqp.Error); ok {
				if amqpErr.Code == amqp.PreconditionFailed {
					// AMQP spec dictates that precondition-failed is returned
					// if the queue has pending messages or active consumers.
					// RabbitMQ, however, does not indicate an error here.
					channel = nil
					continue
				}
			}

			panic(err)
			// return err
		}
	}

	return nil
}
