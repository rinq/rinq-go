package notify

import (
	"github.com/over-pass/overpass-go/src/internals/service"
	"github.com/over-pass/overpass-go/src/overpass"
)

// Listener accepts notifications sent by a notifier.
type Listener interface {
	service.Service

	Listen(id overpass.SessionID, handler overpass.NotificationHandler) (bool, error)
	Unlisten(id overpass.SessionID) (bool, error)
}
