package service

import "sync"

// Link creates bidirectional links between the given services, such that when
// any one service stops, all the others are stopped.
func Link(services ...Service) {
	var once sync.Once

	for i, s := range services {
		go func(i int, s Service) {
			<-s.Done()
			once.Do(func() {
				for j, x := range services {
					if i != j {
						x.Stop()
					}
				}
			})
		}(i, s)
	}
}
