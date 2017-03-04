package notify

import (
	"context"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// Notifier is a low-level interface for sending notifications.
type Notifier interface {
	// NotifyUnicast sends a notification to a specific session.
	NotifyUnicast(
		ctx context.Context,
		msgID ident.MessageID,
		s ident.SessionID,
		t string,
		out *rinq.Payload,
	) (string, error)

	// NotifyMulticast sends a notification to all sessions matching a constraint.
	NotifyMulticast(
		ctx context.Context,
		msgID ident.MessageID,
		c rinq.Constraint,
		t string,
		out *rinq.Payload,
	) (string, error)
}
