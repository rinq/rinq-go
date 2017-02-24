package service

import (
	"sync"

	"github.com/over-pass/overpass-go/src/internals/deferutil"
	"github.com/over-pass/overpass-go/src/internals/reflectutil"
)

type impl struct {
	// mutex guards all other fields, including stoppingMutex
	mutex sync.RWMutex

	// stoppingMutex is held locked from the time Closer.[Graceful]Stop() is
	// called until Closer.Close() is called. It can be acquired for read by
	// Closer.Lock() to prevent goroutines from starting new work while stopping.
	stoppingMutex sync.RWMutex

	isStopping bool
	isGraceful bool
	isStopped  bool

	done chan struct{}
	stop chan struct{}
	err  error
}

// NewImpl returns a service implementation that can be embedded in a struct
// to provide a standard implementation of the Service interface.
func NewImpl() (Service, *Closer) {
	s := &impl{
		done: make(chan struct{}),
		stop: make(chan struct{}),
	}

	return s, &Closer{impl: s}
}

// Done returns a channel that is closed when the session is closed.
func (s *impl) Done() <-chan struct{} {
	return s.done
}

// Err returns the error that caused the Done() channel to close, if any.
func (s *impl) Err() error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.err
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
	unlock := deferutil.Lock(&s.mutex)
	defer unlock()

	if s.isStopped {
		return s.err
	}

	if !s.isStopping {
		s.stoppingMutex.Lock()
		s.isStopping = true
		s.isGraceful = isGraceful
		close(s.stop)
	}

	unlock()

	<-s.done

	return s.Err()
}

// Closer is used to wait for a stop signal and close a service.
type Closer struct {
	impl *impl
}

// Lock acquires a read-lock on the service, preventing it from being stopped
// while the lock is held.
//
// When called, u unlocks the lock.
//
// ok is true if the lock is acquired, or false if the service is already
// stopping.
func (c *Closer) Lock() (u func(), ok bool) {
	c.impl.mutex.RLock()

	if c.impl.isStopping || c.impl.isStopped {
		c.impl.mutex.RUnlock()
		return nil, false
	}

	c.impl.stoppingMutex.RLock()

	locked := true
	return func() {
		if locked {
			c.impl.mutex.RUnlock()
			c.impl.stoppingMutex.RUnlock()
			locked = false
		}
	}, true
}

// Stop returns a channel that is closed the first time service.Stop() or
// service.GracefulStop() is called.
func (c *Closer) Stop() <-chan struct{} {
	return c.impl.stop
}

// IsGraceful returns true if the service.GracefulStop() was called.
func (c *Closer) IsGraceful() bool {
	c.impl.mutex.RLock()
	defer c.impl.mutex.RUnlock()

	return c.impl.isGraceful
}

// Close marks the service as stopped. It must be called to finalize the service
// whether [Graceful]Stop() was called or not.
func (c *Closer) Close(err error) {
	c.impl.mutex.Lock()
	defer c.impl.mutex.Unlock()

	if c.impl.isStopped {
		return
	}

	if !reflectutil.IsNil(err) {
		c.impl.err = err
	}

	c.impl.isStopped = true
	close(c.impl.done)

	if c.impl.isStopping {
		c.impl.stoppingMutex.Unlock()
	}
}
