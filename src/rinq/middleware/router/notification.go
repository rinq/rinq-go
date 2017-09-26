package router

import (
	"context"
	"sync"

	"github.com/rinq/rinq-go/src/rinq"
)

// NotificationRouter dispatches a notification to a specific handler based on
// its type.
type NotificationRouter struct {
	mutex    sync.RWMutex
	handlers map[string]rinq.NotificationHandler
}

// Add routes notifications of type t to h.
func (d *NotificationRouter) Add(t string, h rinq.NotificationHandler) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.handlers == nil {
		d.handlers = map[string]rinq.NotificationHandler{}
	}

	d.handlers[t] = h
}

// Remove removes an existing route for the cmd command.
func (d *NotificationRouter) Remove(cmd string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	delete(d.handlers, cmd)
}

// Handle accepts an incoming notification and routes it to the appropriate handler.
func (d *NotificationRouter) Handle(ctx context.Context, target rinq.Session, n rinq.Notification) {
	d.mutex.RLock()
	h, ok := d.handlers[n.Type]
	d.mutex.RUnlock()

	if ok {
		h(ctx, target, n)
		return
	}

	n.Payload.Close()
}
