package notifications

import (
	"bytes"

	"github.com/rinq/rinq-go/src/internal/notifications"
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

func packCommon(msg *amqp.Publishing, n *notifications.Notification) {
	msg.MessageId = n.ID.String()
	msg.Type = n.Type
	msg.Body = n.Payload.Bytes()

	amqpx.SetHeader(msg, namespaceHeader, n.Namespace)
	marshaling.PackTrace(msg, n.TraceID)
}

func unpackCommon(msg *amqp.Delivery, n *notifications.Notification) error {
	var err error
	n.ID, err = ident.ParseMessageID(msg.MessageId)
	if err != nil {
		return err
	}

	n.Namespace, err = amqpx.GetHeaderString(msg, namespaceHeader)
	if err != nil {
		return err
	}

	n.TraceID = marshaling.UnpackTrace(msg)
	n.Type = msg.Type
	n.Payload = rinq.NewPayloadFromBytes(msg.Body)

	return nil
}

func packUnicastSpecific(msg *amqp.Publishing, n *notifications.Notification) {
	amqpx.SetHeader(msg, targetHeader, n.UnicastTarget.String())
}

func unpackUnicastSpecific(msg *amqp.Delivery, n *notifications.Notification) error {
	t, err := amqpx.GetHeaderString(msg, targetHeader)
	if err != nil {
		return err
	}

	n.UnicastTarget, err = ident.ParseSessionID(t)

	return err
}

func packMulticastSpecific(msg *amqp.Publishing, n *notifications.Notification, buf *bytes.Buffer) {
	cbor.MustEncode(buf, n.MulticastConstraint)
	amqpx.SetHeader(msg, constraintHeader, buf.Bytes())
}

func unpackMulticastSpecific(msg *amqp.Delivery, n *notifications.Notification) error {
	buf, err := amqpx.GetHeaderBytes(msg, constraintHeader)
	if err != nil {
		return err
	}

	return cbor.DecodeBytes(buf, &n.MulticastConstraint)
}
