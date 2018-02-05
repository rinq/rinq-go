package notifications

import (
	"bytes"
	"errors"

	"github.com/rinq/rinq-go/src/internal/notifications"
	"github.com/rinq/rinq-go/src/internal/x/cbor"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
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

func packCommon(msg *amqp.Publishing, n *notifications.Common) {
	msg.MessageId = n.ID.String()
	msg.Type = n.Type
	msg.Body = n.Payload.Bytes()
	msg.Headers = amqp.Table{
		namespaceHeader: n.Namespace,
	}

	marshaling.PackTrace(msg, n.TraceID)
}

func unpackCommon(msg *amqp.Delivery, n *notifications.Common) error {
	var err error
	n.ID, err = ident.ParseMessageID(msg.MessageId)
	if err != nil {
		return err
	}

	n.TraceID = marshaling.UnpackTrace(msg)

	var ok bool
	n.Namespace, ok = msg.Headers[namespaceHeader].(string)
	if !ok {
		return errors.New("namespace header is not a string")
	}

	n.Type = msg.Type
	n.Payload = rinq.NewPayloadFromBytes(msg.Body)

	return nil
}

func packUnicastSpecific(msg *amqp.Publishing, n *notifications.Common) {
	msg.Headers[targetHeader] = n.UnicastTarget.String()
}

func unpackUnicastSpecific(msg *amqp.Delivery, n *notifications.Common) (err error) {
	if t, ok := msg.Headers[targetHeader].(string); ok {
		n.UnicastTarget, err = ident.ParseSessionID(t)
	} else {
		err = errors.New("target header is not a string")
	}

	return
}

func packMulticastSpecific(msg *amqp.Publishing, n *notifications.Common, buf *bytes.Buffer) {
	cbor.MustEncode(buf, n.MulticastConstraint)
	msg.Headers[constraintHeader] = buf.Bytes()
}

func unpackMulticastSpecific(msg *amqp.Delivery, n *notifications.Common) (err error) {
	if buf, ok := msg.Headers[constraintHeader].([]byte); ok {
		err = cbor.DecodeBytes(buf, &n.MulticastConstraint)
	} else {
		err = errors.New("constraint header is not a byte slice")
	}

	return
}
