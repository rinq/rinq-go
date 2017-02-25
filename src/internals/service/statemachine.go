package service

import (
	"sync"

	"github.com/over-pass/overpass-go/src/internals/deferutil"
	"github.com/over-pass/overpass-go/src/internals/reflectutil"
)

// State is a handler for a particular application state.
//
// The function should block until a state transition is necessary.
//
// If next is nil, the service stops such that the Done() channel is closed and
// s.Err() returns err, which may be nil.
//
// Otherwise, next is called and the process is repeated.
type State func() (next State, err error)

// StateMachine is a state-machine based implementation of the Service interface.
type StateMachine struct {
	Forceful  chan struct{}
	Graceful  chan struct{}
	Finalized chan struct{}

	mutex sync.RWMutex
	err   error
}

// NewStateMachine returns a new service trait.
func NewStateMachine() *StateMachine {
	return &StateMachine{
		Forceful:  make(chan struct{}),
		Graceful:  make(chan struct{}),
		Finalized: make(chan struct{}),
	}
}

// Run enters the initial state and runs until the service stops.
func (s *StateMachine) Run(st State) error {
	var err error

	for st != nil && err == nil {
		st, err = st()
	}

	s.finalize(err)

	return err
}

// Done returns a channel that is closed when the session is closed.
func (s *StateMachine) Done() <-chan struct{} {
	return s.Finalized
}

// Err returns the error that caused the Done() channel to close, if any.
func (s *StateMachine) Err() error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.err
}

// Stop halts the service immediately.
func (s *StateMachine) Stop() error {
	unlock := deferutil.Lock(&s.mutex)
	defer unlock()

	select {
	case <-s.Finalized:
		return s.err
	default:
		close(s.Forceful)
	}

	unlock()

	<-s.Finalized

	return s.Err()
}

// GracefulStop halts the service once it has finished any pending work.
func (s *StateMachine) GracefulStop() error {
	unlock := deferutil.Lock(&s.mutex)
	defer unlock()

	select {
	case <-s.Forceful:
		return s.err
	case <-s.Graceful:
		return s.err
	default:
		close(s.Graceful)
	}

	unlock()

	<-s.Finalized

	return s.Err()
}

// finalize should be called after the service is stopped, regardless of whether
// the stop was requested by [Graceful]Stop().
func (s *StateMachine) finalize(err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	select {
	case <-s.Finalized:
		return
	default:
	}

	if !reflectutil.IsNil(err) {
		s.err = err
	}

	close(s.Finalized)

	select {
	case <-s.Forceful:
	default:
		close(s.Forceful)
	}

	select {
	case <-s.Graceful:
	default:
		close(s.Graceful)
	}
}
