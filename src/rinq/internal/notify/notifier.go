package notify

import (
	"context"

	"github.com/rinq/rinq-go/src/rinq"
)

// Notifier is a low-level interface for sending notifications.
type Notifier interface {
	// NotifyUnicast sends a notification to a specific session.
	NotifyUnicast(
		ctx context.Context,
		msgID rinq.MessageID,
		sessID rinq.SessionID,
		notificationType string,
		payload *rinq.Payload,
	) (string, error)

	// NotifyMulticast sends a notification to all sessions matching a constraint.
	NotifyMulticast(
		ctx context.Context,
		msgID rinq.MessageID,
		constraint rinq.Constraint,
		notificationType string,
		payload *rinq.Payload,
	) (string, error)
}
