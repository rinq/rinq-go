package commandamqp

import (
	"sync"

	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/streadway/amqp"
)

// balancedRequestQueue returns the name of the queue used for balanced
// command requests in the given namespace.
func balancedRequestQueue(namespace string) string {
	return "cmd." + namespace
}

// requestQueue returns the name of the queue used for unicast and multicast
// command requests.
func requestQueue(id overpass.PeerID) string {
	return id.ShortString() + ".req"
}

// responseQueue returns the name of the queue used for command responses.
func responseQueue(id overpass.PeerID) string {
	return id.ShortString() + ".rsp"
}

// queueSet declares AMQP resources for queuing balanced command requests.
type queueSet struct {
	mutex  sync.Mutex
	queues map[string]string
}

// Get declares the AMQP queue used for balanced command requests in the given
// namespace and returns the queue name.
func (s *queueSet) Get(channel *amqp.Channel, namespace string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if queue, ok := s.queues[namespace]; ok {
		return queue, nil
	}

	queue := balancedRequestQueue(namespace)

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
