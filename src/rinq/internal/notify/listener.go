package notify

import (
	"context"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/service"
)

// NotificationHandler is a callback-function invoked when a inter-session
// notification is received.
type NotificationHandler func(
	ctx context.Context,
	msgID ident.MessageID,
	target rinq.Session,
	n rinq.Notification,
)

// Listener accepts notifications sent by a notifier.
type Listener interface {
	service.Service

	Listen(id ident.SessionID, ns string, h NotificationHandler) (bool, error)
	Unlisten(id ident.SessionID, ns string) (bool, error)
	UnlistenAll(id ident.SessionID) error
}
