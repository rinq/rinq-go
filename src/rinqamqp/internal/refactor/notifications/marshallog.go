package notifications

import (
	"github.com/jmalloc/twelf/src/twelf"
	"github.com/rinq/rinq-go/src/internal/notifications"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinqamqp/internal/refactor/marshaling"
	"github.com/streadway/amqp"
)

func logSpanMarshalError(l twelf.Logger, n *notifications.Notification, err error) {
	l.Debug(
		"%s ignored outbound span context: %s [%s]",
		n.ID.ShortString(),
		err,
		n.TraceID,
	)
}

func logSpanUnmarshalError(l twelf.Logger, p ident.PeerID, msg *amqp.Delivery, err error) {
	l.Debug(
		"%s ignored inbound span context in %s: %s [%s]",
		p.ShortString(),
		msg.MessageId,
		err,
		marshaling.UnpackTrace(msg),
	)
}
