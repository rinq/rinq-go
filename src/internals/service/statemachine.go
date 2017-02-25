package service

import "sync"

// State is a handler for a particular application state.
//
// The function should block until a state transition is necessary.
//
// If next is nil, the service stops such that the Done() channel is closed and
// s.Err() returns err, which may be nil.
//
// Otherwise, next is called and the process is repeated.
type State func() (next State, err error)

// Finalizer is called when the state machine stops.
type Finalizer func(error) error

// StateMachine is a state-machine based implementation of the Service interface.
type StateMachine struct {
	Forceful  chan struct{}
	Graceful  chan struct{}
	Finalized chan struct{}

	state     State
	finalizer Finalizer

	mutex sync.RWMutex
	err   error
}

// NewStateMachine returns a new service trait.
func NewStateMachine(
	s State,
	f Finalizer,
) *StateMachine {
	return &StateMachine{
		Forceful:  make(chan struct{}),
		Graceful:  make(chan struct{}),
		Finalized: make(chan struct{}),

		state:     s,
		finalizer: f,
	}
}

// Run enters the initial state and runs until the service stops.
func (s *StateMachine) Run() {
	var err error

	for s.state != nil && err == nil {
		s.state, err = s.state()
	}

	err = s.finalizer(err)

	s.mutex.Lock()
	s.err = err
	s.mutex.Unlock()

	s.close()
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
func (s *StateMachine) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	select {
	case <-s.Finalized:
		return
	default:
		close(s.Forceful)
	}
}

// GracefulStop halts the service once it has finished any pending work.
func (s *StateMachine) GracefulStop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	select {
	case <-s.Forceful:
		return
	case <-s.Graceful:
		return
	default:
		close(s.Graceful)
	}
}

func (s *StateMachine) close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	select {
	case <-s.Finalized:
		return
	default:
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
