package notifyamqp

import "github.com/rinq/rinq-go/src/rinq"

// notifyQueue returns the name of the queue used for incoming notifications.
func notifyQueue(id rinq.PeerID) string {
	return id.ShortString() + ".ntf"
}
