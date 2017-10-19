package service

import "sync"

// Service is an interface for background tasks that can finish with an error.
type Service interface {
	// Done returns a channel that is closed when the service is stopped.
	Done() <-chan struct{}

	// Err returns the error that caused the Done() channel to close, if any.
	Err() error

	// Stop halts the service immediately.
	Stop()

	// GracefulStop() halts the service once it has finished any pending work.
	GracefulStop()
}

// WaitAll returns a channel that is closed when all of the given services are
// done.
func WaitAll(services ...Service) <-chan struct{} {
	var wg sync.WaitGroup

	for _, s := range services {
		wg.Add(1)
		go func(s Service) {
			for range s.Done() {
			}
			wg.Done()
		}(s)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	return done
}
