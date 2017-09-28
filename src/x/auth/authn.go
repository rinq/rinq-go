package auth

import (
	"context"

	"github.com/rinq/rinq-go/src/rinq"
)

// Actor represents an entity that can perform actions, such as a user of an
// application.
type Actor struct {
	Type       string
	Identity   string
	Privileges map[string]struct{}
}

// Authenticator identifies an actor by verifying credentials.
type Authenticator interface {
	Authenticate(ctx context.Context, creds map[string]string) (Actor, error)
}

// Predicate is a function used to check if an actor meets some kind of requirements.
type Predicate func(Actor) bool

// factory is authentication middleware for Rinq commands and notifications.
type factory struct {
	authn Authenticator
}

// func (f *factory) Command(h rinq.CommandHandler, r ...Predicate) rinq.CommandHandler {
// 	return func(ctx context.Context, req rinq.CommandHandler, res rinq.Response) {
// 	}
// }
//
// func (f *factory) Notification(h rinq.NotificationHandler, r ...Predicate) rinq.NotificationHandler {
// 	return func(ctx context.Context, req rinq.CommandHandler, res rinq.Response) {
//
// 	}
// }

func actorFromSession(ctx context.Context, rev rinq.Revision) (a Actor, err error) {
	rev.GetMany(ctx, "rinq.x.authn.customer")
	return
}

// // NewAuthN returns a Rinq command handler that authenticates an actor before
// // dispatching to the next command handler.
// func NewAuthN(a Authenticator, h rinq.CommandHandler) rinq.CommandHandler {
// }
