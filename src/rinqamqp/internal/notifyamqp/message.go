package notifyamqp

import (
	"errors"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/rinq/rinq-go/src/internal/opentr"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/constraint"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/amqputil"
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

func packConstraint(msg *amqp.Publishing, con constraint.Constraint) {
	if msg.Headers == nil {
		msg.Headers = amqp.Table{}
	}

	// don't close p, as it's internal buffer is retained inside the msg header
	msg.Headers[constraintHeader] = rinq.NewPayload(con).Bytes()
}

func unpackConstraint(msg *amqp.Delivery) (con constraint.Constraint, err error) {
	if buf, ok := msg.Headers[constraintHeader].([]byte); ok {
		p := rinq.NewPayloadFromBytes(buf)
		defer p.Close()

		err = p.Decode(&con)
	} else {
		err = errors.New("constraint header is not a byte slice")
	}

	return
}

func unpackSpanOptions(msg *amqp.Delivery, t opentracing.Tracer) (opts []opentracing.StartSpanOption, err error) {
	sc, err := amqputil.UnpackSpanContext(msg, t)

	if err == nil {
		opts = append(opts, opentr.CommonSpanOptions...)
		opts = append(opts, ext.SpanKindConsumer)

		if sc != nil {
			opts = append(opts, opentracing.FollowsFrom(sc))
		}
	}

	return
}
