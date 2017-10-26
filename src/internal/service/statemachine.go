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
	Commands  chan request

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
		Commands:  make(chan request),
		finalizer: f,
	}
}

// Run enters the initial state and runs until the service stops.
func (s *StateMachine) Run() {
	var err error

	for s.state != nil && err == nil {
		s.state, err = s.state()
	}

	if s.finalizer != nil {
		err = s.finalizer(err)
	}

	s.mutex.Lock()
	s.err = err
	s.mutex.Unlock()

	s.close()
}

// Done returns a channel that is closed when the service is stopped.
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
	case <-s.Forceful:
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

type request struct {
	fn    func() error
	reply chan<- error
}

// Do enqueues fn in the command channel to be processed by the state-machine.
// It returns ErrStopped if the state-machine is stopping or has already stopped.
func (s *StateMachine) Do(fn func() error) error {
	reply := make(chan error, 1)
	req := request{fn, reply}

	select {
	case s.Commands <- req:
	case <-s.Graceful:
		return ErrStopped
	case <-s.Forceful:
		return ErrStopped
	case <-s.Finalized:
		return ErrStopped
	}

	select {
	case err := <-reply:
		return err
	case <-s.Graceful:
		return ErrStopped
	case <-s.Forceful:
		return ErrStopped
	case <-s.Finalized:
		return ErrStopped
	}
}

// DoGraceful enqueues fn in the command channel to be processed by the
// state-machine. It differs from s.Call() in that it still enqueues the command
// if the state-machine is stopping gracefully (but not if it has been stopped)
// forcefully.
func (s *StateMachine) DoGraceful(fn func() error) error {
	reply := make(chan error, 1)
	req := request{fn, reply}

	select {
	case s.Commands <- req:
	case <-s.Forceful:
		return ErrStopped
	case <-s.Finalized:
		return ErrStopped
	}

	select {
	case err := <-reply:
		return err
	case <-s.Forceful:
		return ErrStopped
	case <-s.Finalized:
		return ErrStopped
	}
}

// Execute handles a command request.
func (s *StateMachine) Execute(req request) {
	req.reply <- req.fn()
}

func (s *StateMachine) close() {
	// protect against panic() that could occur when closing already closed
	// channels if s.close() were to be called concurrently.
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
