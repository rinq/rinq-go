package notifyamqp

import (
	"errors"
	"fmt"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/amqp/internal/amqputil"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/traceutil"
	"github.com/streadway/amqp"
)

const (
	// namespaceHeader specifies the namespace in notifications
	namespaceHeader = "n"

	// targetHeader specifies the target session for unicast notifications
	targetHeader = "t"

	// constraintHeader specifies the constraint for multicast notifications.
	constraintHeader = "c"
)

func unicastRoutingKey(ns string, p ident.PeerID) string {
	return ns + "." + p.String()
}

func packCommonAttributes(
	msg *amqp.Publishing,
	ns string,
	t string,
	p *rinq.Payload,
) {
	msg.Type = t
	msg.Body = p.Bytes()

	if msg.Headers == nil {
		msg.Headers = amqp.Table{}
	}

	msg.Headers[namespaceHeader] = ns
}

func unpackCommonAttributes(msg *amqp.Delivery) (ns, t string, p *rinq.Payload, err error) {
	t = msg.Type
	p = rinq.NewPayloadFromBytes(msg.Body)

	ns, ok := msg.Headers[namespaceHeader].(string)
	if !ok {
		err = errors.New("namespace header is not a string")
	}

	return
}

func packTarget(msg *amqp.Publishing, target ident.SessionID) {
	if msg.Headers == nil {
		msg.Headers = amqp.Table{}
	}

	msg.Headers[targetHeader] = target.String()
}

func unpackTarget(msg *amqp.Delivery) (id ident.SessionID, err error) {
	if t, ok := msg.Headers[targetHeader].(string); ok {
		id, err = ident.ParseSessionID(t)
	} else {
		err = errors.New("target header is not a string")
	}

	return
}

func packConstraint(msg *amqp.Publishing, con rinq.Constraint) {
	t := amqp.Table{}
	for key, value := range con {
		t[key] = value
	}

	if msg.Headers == nil {
		msg.Headers = amqp.Table{}
	}

	msg.Headers[constraintHeader] = t
}

func unpackConstraint(msg *amqp.Delivery) (rinq.Constraint, error) {
	if t, ok := msg.Headers[constraintHeader].(amqp.Table); ok {
		con := rinq.Constraint{}

		for key, value := range t {
			if v, ok := value.(string); ok {
				con[key] = v
			} else {
				return nil, fmt.Errorf("constraint key %s contains non-string value", key)
			}
		}

		return con, nil
	}

	return nil, errors.New("constraint header is not a table")
}

func unpackSpanOptions(msg *amqp.Delivery, t opentracing.Tracer) (opts []opentracing.StartSpanOption, err error) {
	sc, err := amqputil.UnpackSpanContext(msg, t)

	if err == nil {
		opts = append(opts, traceutil.CommonSpanOptions...)
		opts = append(opts, ext.SpanKindConsumer)

		if sc != nil {
			opts = append(opts, opentracing.FollowsFrom(sc))
		}
	}

	return
}
