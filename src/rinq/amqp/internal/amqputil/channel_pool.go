package amqputil

import (
	"context"
	"errors"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/internal/service"
	"github.com/streadway/amqp"
)

const maxPreFetch = ^uint(0) >> 1 // largest int value as uint

// ChannelPool provides a pool of reusable AMQP channels.
type ChannelPool interface {
	service.Service

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
func NewChannelPool(
	broker *amqp.Connection,
	size uint,
	logger rinq.Logger,
) ChannelPool {
	p := &channelPool{
		broker:     broker,
		get:        make(chan getRequest),
		put:        make(chan *amqp.Channel),
		amqpClosed: make(chan *amqp.Error, 1),
		logger:     logger,

		// TODO: make channels into a stack like slice
		channels: make(chan *amqp.Channel, size),
	}

	p.sm = service.NewStateMachine(p.run, p.finalize)
	p.Service = p.sm

	p.broker.NotifyClose(p.amqpClosed)

	go p.sm.Run()

	return p
}

type channelPool struct {
	service.Service
	sm *service.StateMachine

	broker     *amqp.Connection
	get        chan getRequest
	put        chan *amqp.Channel
	amqpClosed chan *amqp.Error
	logger     rinq.Logger

	// TODO: make channels into a stack like slice
	// state-machine data
	channels chan *amqp.Channel
}

type getRequest struct {
	reply chan getResponse
}

type getResponse struct {
	channel *amqp.Channel
	err     error
}

func (p *channelPool) Get() (channel *amqp.Channel, err error) {
	request := getRequest{make(chan getResponse, 1)}

	select {
	case p.get <- request:
		response := <-request.reply
		return response.channel, response.err
	case <-p.sm.Graceful:
		return nil, context.Canceled
	case <-p.sm.Forceful:
		return nil, context.Canceled
	}
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
	p.put <- channel
}

func (p *channelPool) handleGet(request getRequest) (service.State, error) {
	var response getResponse
	select {
	case response.channel = <-p.channels: // fetch from the pool
	default: // none available, make a new channel
		response.channel, response.err = p.broker.Channel()
	}
	request.reply <- response

	return nil, response.err
}

func (p *channelPool) handlePut(channel *amqp.Channel) (service.State, error) {
	if channel == nil {
		return nil, nil
	}

	// set the QoS state back to unlimited, both to "reset" the channel, and to
	// verify that it is still usable.
	if err := channel.Qos(0, 0, true); err != nil {
		return nil, err
	}

	select {
	case p.channels <- channel: // return to the pool
	default: // pool is full, close channel
		_ = channel.Close()
	}

	return nil, nil
}

func (p *channelPool) run() (service.State, error) {
	logChannelPoolStart(p.logger)

	for {
		select {
		case request := <-p.get:
			p.handleGet(request)

		case channel := <-p.put:
			p.handlePut(channel)

		// TODO: cleanupTick thing
		// case <-p.nextCleanupTick

		case <-p.sm.Graceful:
			return nil, nil

		case <-p.sm.Forceful:
			return nil, nil

		case err := <-p.amqpClosed:
			return nil, err
		}
	}
}

func (p *channelPool) finalize(err error) error {
	logChannelPoolStop(p.logger, err)
	return err
}
