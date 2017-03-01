package syncutil

import "sync"

// Group returns a channel that is closed when all of the given channels are
// closed.
func Group(chans ...<-chan struct{}) <-chan struct{} {
	var wg sync.WaitGroup

	for _, c := range chans {
		wg.Add(1)
		go func(c <-chan struct{}) {
			for range c {
			}
			wg.Done()
		}(c)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	return done
}
