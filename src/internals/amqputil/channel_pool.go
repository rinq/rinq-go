package amqputil

import "github.com/streadway/amqp"

// ChannelPool provides a pool of reusable AMQP channels.
type ChannelPool interface {
	// Get fetches a channel from the pool, or creates one as necessary.
	Get() (*amqp.Channel, error)

	// Put returns a channel to the pool.
	Put(*amqp.Channel)
}

// NewChannelPool returns a channel pool of the given size.
func NewChannelPool(broker *amqp.Connection, size uint) ChannelPool {
	return &channelPool{
		broker:   broker,
		channels: make(chan *amqp.Channel, size),
	}
}

type channelPool struct {
	broker   *amqp.Connection
	channels chan *amqp.Channel
}

func (p *channelPool) Get() (channel *amqp.Channel, err error) {
	select {
	case channel = <-p.channels: // fetch from the pool
	default: // none available, make a new channel
		channel, err = p.broker.Channel()
	}

	return
}

func (p *channelPool) Put(channel *amqp.Channel) {
	if channel == nil {
		return
	}

	// set the QoS state back to unlimited, both to "reset" the channel, and to
	// verify that it is still usable.
	if err := channel.Qos(0, 0, true); err != nil {
		return
	}

	select {
	case p.channels <- channel: // return to the pool
	default: // pool is full, close channel
		channel.Close()
	}
}
