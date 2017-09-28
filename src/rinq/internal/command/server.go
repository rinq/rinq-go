package command

import (
	"context"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/service"
)

// Handler is a callback-function invoked when a command request is
// received by the peer.
type Handler func(
	ctx context.Context,
	msgID ident.MessageID,
	req rinq.Request,
	res rinq.Response,
)

// Server processes command requests made by an invoker.
type Server interface {
	service.Service

	Listen(namespace string, handler Handler) (bool, error)
	Unlisten(namespace string) (bool, error)
}
