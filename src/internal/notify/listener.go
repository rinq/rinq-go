package notify

import (
	"github.com/rinq/rinq-go/src/internal/service"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// Listener accepts notifications sent by a notifier.
type Listener interface {
	service.Service

	Listen(id ident.SessionID, ns string, h rinq.NotificationHandler) (bool, error)
	Unlisten(id ident.SessionID, ns string) (bool, error)
	UnlistenAll(id ident.SessionID) error
}
