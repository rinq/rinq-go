package internals

import (
	"context"

	"github.com/over-pass/overpass-go/src/overpass"
)

// Notifier is a low-level interface for sending notifications.
type Notifier interface {
	// NotifyUnicast sends a notification to a specific session.
	NotifyUnicast(
		ctx context.Context,
		msgID overpass.MessageID,
		sessID overpass.SessionID,
		notificationType string,
		payload *overpass.Payload,
	) error

	// NotifyMulticast sends a notification to all sessions matching a constraint.
	NotifyMulticast(
		ctx context.Context,
		msgID overpass.MessageID,
		constraint overpass.Constraint,
		notificationType string,
		payload *overpass.Payload,
	) error
}
