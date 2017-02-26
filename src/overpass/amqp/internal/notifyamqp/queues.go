package notifyamqp

import "github.com/over-pass/overpass-go/src/overpass"

// notifyQueue returns the name of the queue used for incoming notifications.
func notifyQueue(id overpass.PeerID) string {
	return id.ShortString() + ".ntf"
}
