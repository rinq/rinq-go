package amqpx

import (
	"github.com/streadway/amqp"
)

// ChannelWithPreFetch returns new channel with a channel-wide pre-fetch limit.
func ChannelWithPreFetch(broker *amqp.Connection, preFetch int) (*amqp.Channel, error) {
	// Always use a "channel-wide" QoS setting.
	// http://www.rabbitmq.com/consumer-prefetch.html
	caps, _ := broker.Properties["capabilities"].(amqp.Table)
	global, _ := caps["per_consumer_qos"].(bool)

	c, err := broker.Channel()
	if err != nil {
		return nil, err
	}

	err = c.Qos(preFetch, 0, global)
	if err != nil {
		_ = c.Close()
		return nil, err
	}

	return c, nil
}
