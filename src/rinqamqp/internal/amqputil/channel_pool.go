package amqputil

import (
	"context"
	"errors"
	"time"

	"github.com/rinq/rinq-go/src/internal/service"
	"github.com/rinq/rinq-go/src/rinq"
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
	Put(*amqp.Channel) error
}

// NewChannelPool returns a channel pool of the given size.
func NewChannelPool(
	broker *amqp.Connection,
	size uint,
	logger rinq.Logger,
) ChannelPool {
	duration := 60 * time.Second
	p := &channelPool{
		broker:     broker,
		get:        make(chan getRequest),
		put:        make(chan *amqp.Channel),
		amqpClosed: make(chan *amqp.Error, 1),
		logger:     logger,

		channels:        make([]*amqp.Channel, 0, size),
		cleanupDuration: duration,
		cleanupTimer:    time.NewTimer(duration),
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

	// state-machine data
	channels        []*amqp.Channel
	cleanupDuration time.Duration
	cleanupTimer    *time.Timer
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
		err = errors.New("pre-fetch is too large")
		logChannelPoolGetQOS(p.logger, len(p.channels), err)
		return nil, err
	}

	err = channel.Qos(int(preFetch), 0, global)
	if err != nil {
		logChannelPoolGetQOS(p.logger, len(p.channels), err)
		return nil, err
	}

	return channel, nil
}

func (p *channelPool) Put(channel *amqp.Channel) error {
	select {
	case p.put <- channel:
		return nil
	case <-p.sm.Forceful:
		return context.Canceled
	}
}

func (p *channelPool) handleGet(request getRequest) error {
	var response getResponse

	index := len(p.channels) - 1
	if index >= 0 {
		// fetch from the pool
		response.channel = p.channels[index]
		p.channels = p.channels[:index]
	} else {
		// none available, make a new channel
		response.channel, response.err = p.broker.Channel()
	}
	request.reply <- response

	if response.err != nil {
		logChannelPoolGet(p.logger, len(p.channels), response.err)
	}

	return response.err
}

func (p *channelPool) handlePut(channel *amqp.Channel) error {
	if channel == nil {
		return nil
	}

	// stop cleanup timer
	if !p.cleanupTimer.Stop() {
		<-p.cleanupTimer.C
	}

	// set the QoS state back to unlimited, both to "reset" the channel, and to
	// verify that it is still usable.
	if err := channel.Qos(0, 0, true); err != nil {
		logChannelPoolPut(p.logger, len(p.channels), err)
		return err
	}

	if len(p.channels) < cap(p.channels) {
		// return to the pool
		p.channels = append(p.channels, channel)
	} else {
		// pool is full, close channel
		if err := channel.Close(); err != nil {
			logChannelPoolPut(p.logger, len(p.channels), err)
			return err
		}
	}

	// restart cleanup timer
	p.cleanupTimer.Reset(p.cleanupDuration)

	return nil
}

func (p *channelPool) handlePeriodicCleanup() error {
	index := len(p.channels) - 1
	if index >= 0 {
		// fetch from the pool
		channel := p.channels[index]
		p.channels = p.channels[:index]
		// close channel
		err := channel.Close()
		logChannelPoolCleanup(p.logger, len(p.channels), err)
		if err != nil {
			return err
		}
	}

	// restart cleanup timer
	p.cleanupTimer.Reset(p.cleanupDuration)

	return nil
}

func (p *channelPool) run() (service.State, error) {
	logChannelPoolStart(p.logger, cap(p.channels))

	for {
		select {
		case request := <-p.get:
			if err := p.handleGet(request); err != nil {
				return nil, err
			}

		case channel := <-p.put:
			if err := p.handlePut(channel); err != nil {
				return nil, err
			}

		case <-p.cleanupTimer.C:
			if err := p.handlePeriodicCleanup(); err != nil {
				return nil, err
			}

		case <-p.sm.Graceful:
			logChannelPoolGraceful(p.logger, len(p.channels))

		case <-p.sm.Forceful:
			return nil, nil

		case err := <-p.amqpClosed:
			return nil, err
		}
	}
}

func (p *channelPool) finalize(err error) error {
	logChannelPoolStop(p.logger, len(p.channels), err)
	return err
}
