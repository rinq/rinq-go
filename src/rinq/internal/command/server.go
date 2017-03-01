package command

import (
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/internal/service"
)

// Server processes command requests made by an invoker.
type Server interface {
	service.Service

	Listen(namespace string, handler rinq.CommandHandler) (bool, error)
	Unlisten(namespace string) (bool, error)
}
