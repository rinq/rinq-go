package internals

import "github.com/over-pass/overpass-go/src/overpass"

// Listener accepts notifications sent by a notifier.
type Listener interface {
	Service

	Listen(id overpass.SessionID, handler overpass.NotificationHandler) error
	Unlisten(id overpass.SessionID) error
}
