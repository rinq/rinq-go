package notify

import (
	"context"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/constraint"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// Notifier is a low-level interface for sending notifications.
type Notifier interface {
	// NotifyUnicast sends a notification to a specific session.
	NotifyUnicast(
		ctx context.Context,
		msgID ident.MessageID,
		traceID string,
		s ident.SessionID,
		ns string,
		t string,
		out *rinq.Payload,
	) error

	// NotifyMulticast sends a notification to all sessions matching a constraint.
	NotifyMulticast(
		ctx context.Context,
		msgID ident.MessageID,
		traceID string,
		con constraint.Constraint,
		ns string,
		t string,
		out *rinq.Payload,
	) error
}
