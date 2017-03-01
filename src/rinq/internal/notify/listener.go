package notify

import (
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/internal/service"
)

// Listener accepts notifications sent by a notifier.
type Listener interface {
	service.Service

	Listen(id rinq.SessionID, handler rinq.NotificationHandler) (bool, error)
	Unlisten(id rinq.SessionID) (bool, error)
}
