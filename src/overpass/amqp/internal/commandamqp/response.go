package commandamqp

import (
	"context"
	"sync"

	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpass/amqp/internal/amqputil"
	"github.com/streadway/amqp"
)

// response is used to send responses to command requests, it implements
// overpass.Response.
type response struct {
	context  context.Context
	channels amqputil.ChannelPool
	msgID    overpass.MessageID

	mutex      sync.RWMutex
	isRequired bool
	isClosed   bool
}

func newResponse(
	ctx context.Context,
	channels amqputil.ChannelPool,
	msgID overpass.MessageID,
	isRequired bool,
) (overpass.Response, func() bool) {
	r := &response{
		context:    ctx,
		channels:   channels,
		msgID:      msgID,
		isRequired: isRequired,
	}

	return r, r.finalize
}

func (r *response) IsRequired() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.isRequired {
		return false
	}

	select {
	case <-r.context.Done():
		r.isRequired = false
		return false
	default:
		return true
	}
}

func (r *response) IsClosed() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.isClosed
}

func (r *response) Done(payload *overpass.Payload) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.isClosed {
		panic("responder is already closed")
	}

	r.respond(amqp.Publishing{
		Type: successResponse,
		Body: payload.Bytes(),
	})
}

func (r *response) Error(err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.isClosed {
		panic("responder is already closed")
	}

	if f, ok := err.(overpass.Failure); ok {
		if f.Type == "" {
			panic("failure type is empty")
		}

		r.respond(amqp.Publishing{
			Type: failureResponse,
			Headers: amqp.Table{
				failureTypeHeader:    f.Type,
				failureMessageHeader: f.Message,
			},
			Body: f.Payload.Bytes(),
		})
	} else {
		r.respond(amqp.Publishing{
			Type: errorResponse,
			Body: []byte(err.Error()),
		})
	}
}

func (r *response) Fail(failureType, message string) overpass.Failure {
	err := overpass.Failure{Type: failureType, Message: message}
	r.Error(err)
	return err
}

func (r *response) Close() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.isClosed {
		return false
	}

	r.respond(amqp.Publishing{Type: successResponse})

	return true
}

func (r *response) finalize() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.isClosed {
		return true
	}

	r.isClosed = true
	r.isRequired = false

	return false
}

func (r *response) respond(msg amqp.Publishing) {
	if r.isRequired {
		channel, err := r.channels.Get()
		if err != nil {
			panic(err)
		}
		defer r.channels.Put(channel)

		amqputil.PackTrace(r.context, &msg)
		amqputil.PackDeadline(r.context, &msg)

		err = channel.Publish(
			responseExchange,
			r.msgID.String(),
			false, // mandatory,
			false, // immediate,
			msg,
		)
		if err != nil {
			panic(err)
		}
	}

	r.isClosed = true
	r.isRequired = false
}
