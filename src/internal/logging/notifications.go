package logging

import (
	"github.com/jmalloc/twelf/src/twelf"
	"github.com/rinq/rinq-go/src/internal/transport"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// SentNotification logs a message about an outgoing notification.
func SentNotification(logger twelf.Logger, n *transport.Notification) {
	if n.IsMulticast {
		logger.Log(
			"%s sent '%s::%s' notification to sessions matching %s (%d/o) [%s]",
			n.ID.ShortString(),
			n.Namespace,
			n.Type,
			n.MulticastConstraint,
			n.Payload.Len(),
			n.TraceID,
		)
	} else {
		logger.Log(
			"%s sent '%s::%s' notification to %s (%d/o) [%s]",
			n.ID.ShortString(),
			n.Namespace,
			n.Type,
			n.UnicastTarget.ShortString(),
			n.Payload.Len(),
			n.TraceID,
		)
	}
}

// ReceivedNotification logs a message about an incoming notification.
func ReceivedNotification(
	logger twelf.Logger,
	ref ident.Ref,
	n *transport.Notification,
) {
	logger.Log(
		"%s received '%s::%s' notification from %s (%d/i) [%s]",
		ref.ShortString(),
		n.Namespace,
		n.Type,
		n.ID.Ref.ShortString(),
		n.Payload.Len(),
		n.TraceID,
	)
}

// IgnoredNotification logs a message about an incoming notification that has
// been ignored due to an error.
func IgnoredNotification(
	logger twelf.Logger,
	ref ident.Ref,
	n *transport.Notification,
	err error,
) {
	logger.Debug(
		"%s ignored '%s::%s' notification from %s (%d/i) [%s]",
		ref.ShortString(),
		n.Namespace,
		n.Type,
		n.ID.Ref.ShortString(),
		n.Payload.Len(),
		err,
		n.TraceID,
	)
}

// NotificationSubscribe logs a message about a session subscribing to
// notifications.
func NotificationSubscribe(
	logger twelf.Logger,
	ref ident.Ref,
	ns string,
) {
	logger.Debug(
		"%s started listening for notifications in '%s' namespace",
		ref.ShortString(),
		ns,
	)
}

// NotificationUnsubscribe logs a message about a session unsubscribing from
// notifications.
func NotificationUnsubscribe(
	logger twelf.Logger,
	ref ident.Ref,
	ns string,
) {
	logger.Debug(
		"%s stopped listening for notifications in '%s' namespace",
		ref.ShortString(),
		ns,
	)
}
