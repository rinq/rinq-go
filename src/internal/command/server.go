package command

import (
	"github.com/rinq/rinq-go/src/internal/service"
	"github.com/rinq/rinq-go/src/rinq"
)

// Server processes command requests made by an invoker.
type Server interface {
	service.Service

	Listen(ns string, h rinq.CommandHandler) (bool, error)
	Unlisten(ns string) (bool, error)
}
