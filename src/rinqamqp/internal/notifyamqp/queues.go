package notifyamqp

import "github.com/rinq/rinq-go/src/rinq/ident"

// notifyQueue returns the name of the queue used for incoming notifications.
func notifyQueue(id ident.PeerID) string {
	return id.ShortString() + ".ntf"
}
