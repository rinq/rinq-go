package amqputil

import (
	"errors"

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
func NewChannelPool(broker *amqp.Connection, size uint) ChannelPool {
	p := &channelPool{
		broker:     broker,
		get:        make(chan getRequest),
		put:        make(chan *amqp.Channel),
		amqpClosed: make(chan *amqp.Error, 1),

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

	// TODO: make channels into a stack like slice
	// state-machine data
	channels chan *amqp.Channel
}

type getRequest struct {
	reply chan<- getResponse
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
		return response.channel, repsonse.err
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

func (p *channelPool) getHandler(request getRequest) (service.State, error) {
	var response getReposne
	select {
	case channel = <-p.channels: // fetch from the pool
	default: // none available, make a new channel
		response.channel, response.err = p.broker.Channel()
	}
	request.reply <- response
	// if response.err != nil {
	// 	return nil, response.err
	// }

	return nil, response.err
}

func (p *channelPool) run() (service.State, error) {
	// TODO: add logging... fmt.Println("channelpool running")
	for {
		select {
		case request := <-p.get:
			// var response getReposne
			// select {
			// case channel = <-p.channels: // fetch from the pool
			// default: // none available, make a new channel
			// 	response.channel, response.err = p.broker.Channel()
			// }
			// request.reply <- response
			// if response.err != nil {
			// 	return nil, response.err
			// }
			return getHandler(request)

		// case channel := <-p.put:
		// 	if channel == nil {
		// 		return
		// 	}

		// 	// set the QoS state back to unlimited, both to "reset" the channel, and to
		// 	// verify that it is still usable.
		// 	if err := channel.Qos(0, 0, true); err != nil {
		// 		return
		// 	}

		// 	select {
		// 	case p.channels <- channel: // return to the pool
		// 	default: // pool is full, close channel
		// 		_ = channel.Close()
		// 	}

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
	// TODO: add logging... fmt.Println("channelpool stopped", err)
	return err
}
