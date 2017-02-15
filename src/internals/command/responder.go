package command

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/streadway/amqp"
)

// responder is used to send responses to command requests, it implements
// overpass.Responder.
type responder struct {
	channels amqputil.ChannelPool
	context  context.Context
	msgID    overpass.MessageID
	logger   *log.Logger

	mutex      sync.RWMutex
	isRequired bool
	isClosed   bool
}

func (r *responder) IsRequired() bool {
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

func (r *responder) IsClosed() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.isClosed
}

func (r *responder) Done(payload *overpass.Payload) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.isClosed {
		panic("responder is already closed")
	}

	r.close(amqp.Publishing{
		Type: successResponse,
		Body: payload.Bytes(),
	})
}

func (r *responder) Error(err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.isClosed {
		panic("responder is already closed")
	}

	if f, ok := err.(overpass.Failure); ok {
		r.close(amqp.Publishing{
			Type: failureResponse,
			Headers: amqp.Table{
				failureTypeHeader:    f.Type,
				failureMessageHeader: f.Message,
			},
			Body: f.Payload.Bytes(),
		})
	} else {
		r.close(amqp.Publishing{
			Type: errorResponse,
			Body: []byte(err.Error()),
		})
	}
}

func (r *responder) Fail(failureType, message string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.isClosed {
		panic("responder is already closed")
	}

	r.close(amqp.Publishing{
		Type: failureResponse,
		Headers: amqp.Table{
			failureTypeHeader:    failureType,
			failureMessageHeader: message,
		},
	})
}

func (r *responder) Close() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.isClosed {
		r.close(amqp.Publishing{Type: successResponse})
	}
}

func (r *responder) close(msg amqp.Publishing) {
	// TODO log
	fmt.Println(msg)
	if r.isRequired {
		channel, err := r.channels.Get()
		if err != nil {
			panic(err)
		}
		defer r.channels.Put(channel)

		amqputil.PutCorrelationID(r.context, &msg)
		amqputil.PutExpiration(r.context, &msg)

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
