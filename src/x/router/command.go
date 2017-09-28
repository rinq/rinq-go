package router

import (
	"context"
	"errors"
	"sync"

	"github.com/rinq/rinq-go/src/rinq"
)

// CommandRouter dispatches a command request to a specific handler based on
// command name.
type CommandRouter struct {
	mutex    sync.RWMutex
	handlers map[string]rinq.CommandHandler
}

// Add routes command requests for cmd to h.
func (d *CommandRouter) Add(cmd string, h rinq.CommandHandler) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.handlers == nil {
		d.handlers = map[string]rinq.CommandHandler{}
	}

	d.handlers[cmd] = h
}

// Remove removes an existing route for the cmd command.
func (d *CommandRouter) Remove(cmd string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	delete(d.handlers, cmd)
}

// Handle accepts an incoming request and routes it to the appropriate handler.
func (d *CommandRouter) Handle(ctx context.Context, req rinq.Request, res rinq.Response) {
	d.mutex.RLock()
	h, ok := d.handlers[req.Command]
	d.mutex.RUnlock()

	if ok {
		h(ctx, req, res)
		return
	}

	req.Payload.Close()
	res.Error(errors.New("unknown command"))
}
