package notifications

import (
	"bytes"

	"github.com/jmalloc/twelf/src/twelf"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rinq/rinq-go/src/internal/transport"
	"github.com/rinq/rinq-go/src/internal/x/cbor"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/refactor/amqpx"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/refactor/marshaling"
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

// Encoder creates AMQP messages from notifications.
type Encoder struct {
	PeerID ident.PeerID
	Tracer opentracing.Tracer
	Logger twelf.Logger
}

// Marshal returns an AMQP message representing n.
// sb and cb are buffers used to encode the OpenTracing span context and
// multicast notification constraint, respectively. They must remain valid
// until the AMQP message is published.
func (e *Encoder) Marshal(n *transport.Notification, sb, cb *bytes.Buffer) (*amqp.Publishing, error) {
	msg := &amqp.Publishing{
		MessageId: n.ID.String(),
		Type:      n.Type,
		Body:      n.Payload.Bytes(),
	}

	amqpx.SetHeader(msg, namespaceHeader, n.Namespace)
	marshaling.PackTrace(msg, n.TraceID)

	if n.IsMulticast {
		cbor.MustEncode(cb, n.MulticastConstraint)
		amqpx.SetHeader(msg, constraintHeader, cb.Bytes())
	} else {
		amqpx.SetHeader(msg, targetHeader, n.UnicastTarget.String())
	}

	if err := marshaling.PackSpanContext(msg, e.Tracer, n.SpanContext, sb); err != nil {
		logSpanMarshalError(e.Logger, n, err) // log but don't return err
	}

	return msg, nil
}

// Decoder creates notifications from AMQP messages.
type Decoder struct {
	PeerID ident.PeerID
	Tracer opentracing.Tracer
	Logger twelf.Logger
}

// Unmarshal returns a notification based on msg.
func (d *Decoder) Unmarshal(msg *amqp.Delivery) (*transport.Notification, error) {
	n := &transport.Notification{
		TraceID: marshaling.UnpackTrace(msg),
		Type:    msg.Type,
		Payload: rinq.NewPayloadFromBytes(msg.Body),
	}

	var err error
	n.ID, err = ident.ParseMessageID(msg.MessageId)
	if err != nil {
		return nil, err
	}

	n.Namespace, err = amqpx.GetHeaderString(msg, namespaceHeader)
	if err != nil {
		return nil, err
	}

	if msg.Exchange == multicastExchange {
		n.IsMulticast = true

		var buf []byte
		buf, err = amqpx.GetHeaderBytes(msg, constraintHeader)
		if err != nil {
			return nil, err
		}

		if err = cbor.DecodeBytes(buf, &n.MulticastConstraint); err != nil {
			return nil, err
		}
	} else {
		var t string
		t, err = amqpx.GetHeaderString(msg, targetHeader)
		if err != nil {
			return nil, err
		}

		n.UnicastTarget, err = ident.ParseSessionID(t)
		if err != nil {
			return nil, err
		}
	}

	n.SpanContext, err = marshaling.UnpackSpanContext(msg, d.Tracer)
	if err != nil {
		logSpanUnmarshalError(d.Logger, d.PeerID, msg, err)
	}

	return n, nil
}
