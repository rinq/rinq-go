package amqpx

import (
	"errors"

	"github.com/streadway/amqp"
)

const maxPreFetch = ^uint(0) >> 1 // largest int value as uint

// ChannelPool provides a pool of reusable AMQP channels.
type ChannelPool interface {
	// Get fetches a channel from the pool, or creates one as necessary.
	Get() (*amqp.Channel, error)

	// GetQOS fetches a channel from the pool and sets the pre-fetch count
	// before returning it. The pre-fetch is applied to across all consumers on
	// the channel.
	GetQOS(preFetch uint) (*amqp.Channel, error)

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

// GetQOS fetches a channel from the pool and sets the pre-fetch count
// before returning it. The pre-fetch is applied across all consumers on
// the channel.
func (p *channelPool) GetQOS(preFetch uint) (*amqp.Channel, error) {
	channel, err := p.Get()
	if err != nil {
		return nil, err
	}

	// Always use a "channel-wide" QoS setting.
	// http://www.rabbitmq.com/consumer-prefetch.html
	caps, _ := p.broker.Properties["capabilities"].(amqp.Table)
	global, _ := caps["per_consumer_qos"].(bool)

	if preFetch > maxPreFetch {
		return nil, errors.New("pre-fetch is too large")
	}

	err = channel.Qos(int(preFetch), 0, global)
	if err != nil {
		return nil, err
	}

	return channel, nil
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
		_ = channel.Close()
	}
}
