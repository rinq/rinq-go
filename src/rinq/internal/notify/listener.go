package notify

import (
	"context"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/service"
)

// Handler is a callback-function invoked when a inter-session
// notification is received.
type Handler func(
	ctx context.Context,
	msgID ident.MessageID,
	target rinq.Session,
	n rinq.Notification,
)

// Listener accepts notifications sent by a notifier.
type Listener interface {
	service.Service

	Listen(id ident.SessionID, ns string, h Handler) (bool, error)
	Unlisten(id ident.SessionID, ns string) (bool, error)
	UnlistenAll(id ident.SessionID) error
}
