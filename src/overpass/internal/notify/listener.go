package notify

import (
	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpass/internal/service"
)

// Listener accepts notifications sent by a notifier.
type Listener interface {
	service.Service

	Listen(id overpass.SessionID, handler overpass.NotificationHandler) (bool, error)
	Unlisten(id overpass.SessionID) (bool, error)
}
