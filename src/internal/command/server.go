package command

import (
	"github.com/over-pass/overpass-go/src/internal/service"
	"github.com/over-pass/overpass-go/src/overpass"
)

// Server processes command requests made by an invoker.
type Server interface {
	service.Service

	Listen(namespace string, handler overpass.CommandHandler) (bool, error)
	Unlisten(namespace string) (bool, error)
}
