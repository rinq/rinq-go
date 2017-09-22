package commandamqp

import (
	"errors"
	"fmt"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/amqp/internal/amqputil"
	"github.com/rinq/rinq-go/src/rinq/internal/traceutil"
	"github.com/streadway/amqp"
)

const (
	// successResponse is the AMQP message type used for successful call
	// responses.
	successResponse = "s"

	// failureResponse is the AMQP message type used for call responses
	// indicating failure for an "expected" application-defined reason.
	failureResponse = "f"

	// errorResponse is the AMQP message type used for call responses indicating
	// unepected error or internal error.
	errorResponse = "e"
)

const (
	// namespaceHeader specifies the namespace in command requests and
	// uncorrelated command responses.
	namespaceHeader = "n"

	// commandHeader specifies the command name requests and
	// uncorrelated command responses.
	commandHeader = "c"

	// failureTypeHeader specifies the failure type in command responses with
	// the "failureResponse" type.
	failureTypeHeader = "t"

	// failureMessageHeader holds the error message in command responses with
	// the "failureResponse" type.
	failureMessageHeader = "m"
)

type replyMode string

const (
	// replyNone is the AMQP reply-to value used for command requests that are
	// not expecting a reply.
	replyNone replyMode = ""

	// replyNormal is the AMQP reply-to value used for command requests that are
	// waiting for a reply.
	replyCorrelated replyMode = "c"

	// replyUncorrelated is the AMQP reply-to value used for command requests
	// that are waiting for a reply, but where the invoker does not have
	// any information about the request. This instruct the server to include
	// request information in the response.
	replyUncorrelated replyMode = "u"
)

func packNamespaceAndCommand(msg *amqp.Publishing, ns, cmd string) {
	if msg.Headers == nil {
		msg.Headers = amqp.Table{}
	}

	msg.Headers[namespaceHeader] = ns
	msg.Headers[commandHeader] = cmd
}

func unpackNamespaceAndCommand(msg *amqp.Delivery) (ns string, cmd string, err error) {
	ns, ok := msg.Headers[namespaceHeader].(string)
	if !ok {
		err = errors.New("namespace header is not a string")
	}

	cmd, ok = msg.Headers[commandHeader].(string)
	if !ok {
		err = errors.New("command header is not a string")
	}

	return
}

func packReplyMode(msg *amqp.Publishing, m replyMode) {
	msg.ReplyTo = string(m)
}

func unpackReplyMode(msg *amqp.Delivery) replyMode {
	return replyMode(msg.ReplyTo)
}

func packRequest(
	msg *amqp.Publishing,
	ns string,
	cmd string,
	p *rinq.Payload,
	m replyMode,
) {
	packNamespaceAndCommand(msg, ns, cmd)
	packReplyMode(msg, m)
	msg.Body = p.Bytes()
}

func packSuccessResponse(msg *amqp.Publishing, p *rinq.Payload) {
	msg.Type = successResponse
	msg.Body = p.Bytes()
}

func packErrorResponse(msg *amqp.Publishing, err error) {
	if f, ok := err.(rinq.Failure); ok {
		if f.Type == "" {
			panic("failure type is empty")
		}

		msg.Type = failureResponse
		msg.Body = f.Payload.Bytes()

		if msg.Headers == nil {
			msg.Headers = amqp.Table{}
		}

		msg.Headers[failureTypeHeader] = f.Type
		if f.Message != "" {
			msg.Headers[failureMessageHeader] = f.Message
		}

	} else {
		msg.Type = errorResponse
		msg.Body = []byte(err.Error())
	}
}

func unpackResponse(msg *amqp.Delivery) (*rinq.Payload, error) {
	switch msg.Type {
	case successResponse:
		return rinq.NewPayloadFromBytes(msg.Body), nil

	case failureResponse:
		failureType, _ := msg.Headers[failureTypeHeader].(string)
		if failureType == "" {
			return nil, errors.New("malformed response, failure type must be a non-empty string")
		}

		failureMessage, _ := msg.Headers[failureMessageHeader].(string)

		payload := rinq.NewPayloadFromBytes(msg.Body)
		return payload, rinq.Failure{
			Type:    failureType,
			Message: failureMessage,
			Payload: payload,
		}

	case errorResponse:
		return nil, rinq.CommandError(msg.Body)

	default:
		return nil, fmt.Errorf("malformed response, message type '%s' is unexpected", msg.Type)
	}
}

func unpackSpanOptions(msg *amqp.Delivery, t opentracing.Tracer) (opts []opentracing.StartSpanOption, err error) {
	sc, err := amqputil.UnpackSpanContext(msg, t)

	if err == nil {
		opts = append(opts, traceutil.CommonSpanOptions...)
		opts = append(opts, ext.SpanKindRPCServer)

		if sc != nil {
			if unpackReplyMode(msg) == replyCorrelated {
				opts = append(opts, opentracing.ChildOf(sc))
			} else {
				opts = append(opts, opentracing.FollowsFrom(sc))
			}
		}
	}

	return
}
