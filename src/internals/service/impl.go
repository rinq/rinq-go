package service

import (
	"sync"
	"sync/atomic"

	"github.com/over-pass/overpass-go/src/internals/reflectutil"
)

type impl struct {
	once       sync.Once
	done       chan struct{}
	stop       chan struct{}
	err        atomic.Value
	isGraceful atomic.Value
}

// NewImpl returns a service implementation that can be embedded in a struct
// to provide a standard implementation of the Service interface.
func NewImpl() (Service, *Closer) {
	h := &impl{
		done: make(chan struct{}),
		stop: make(chan struct{}),
	}

	return h, &Closer{impl: h}
}

// Done returns a channel that is closed when the session is closed.
func (s *impl) Done() <-chan struct{} {
	return s.done
}

// Err returns the error that caused the Done() channel to close, if any.
func (s *impl) Err() error {
	err, _ := s.err.Load().(error)
	return err
}

// Stop halts the service immediately.
func (s *impl) Stop() error {
	return s.doStop(false)
}

// GracefulStop() halts the service once it has finished any pending work.
func (s *impl) GracefulStop() error {
	return s.doStop(true)
}

func (s *impl) doStop(isGraceful bool) error {
	s.once.Do(func() {
		s.isGraceful.Store(isGraceful)
		close(s.stop)
		<-s.done
	})

	return s.Err()
}

// Closer is used to wait for a stop signal and close a service.
type Closer struct {
	once sync.Once
	impl *impl
}

// Stop returns a channel that is closed the first time service.Stop() or
// service.GracefulStop() is called.
func (c *Closer) Stop() <-chan struct{} {
	return c.impl.stop
}

// IsGraceful returns true if the service.GracefulStop() was called.
func (c *Closer) IsGraceful() bool {
	isGraceful, _ := c.impl.isGraceful.Load().(bool)
	return isGraceful
}

// Close closes the done channel and sets the error, if any.
func (c *Closer) Close(err error) {
	c.once.Do(func() {
		if !reflectutil.IsNil(err) {
			c.impl.err.Store(err)
		}
		close(c.impl.done)
	})
}
