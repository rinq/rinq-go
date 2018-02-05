package amqpx

import (
	"errors"

	"github.com/streadway/amqp"
)

const maxPreFetch = ^uint(0) >> 1 // largest int value as uint

// ChannelWithQOS opens a new channel and sets the pre-fetch count before returning
// it. The pre-fetch is applied across all consumers on the channel.
func ChannelWithQOS(broker *amqp.Connection, preFetch uint) (*amqp.Channel, error) {
	// Always use a "channel-wide" QoS setting.
	// http://www.rabbitmq.com/consumer-prefetch.html
	caps, _ := broker.Properties["capabilities"].(amqp.Table)
	global, _ := caps["per_consumer_qos"].(bool)

	if preFetch > maxPreFetch {
		return nil, errors.New("pre-fetch is too large")
	}

	c, err := broker.Channel()
	if err != nil {
		return nil, err
	}

	err = c.Qos(int(preFetch), 0, global)
	if err != nil {
		_ = c.Close()
		return nil, err
	}

	return c, nil
}
