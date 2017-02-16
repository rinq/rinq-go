package service

// Wait blocks until one of the services fails, or all complete successfully.
// describing the ones that failed.
func Wait(services ...Service) error {
	count := len(services)
	errors := make(chan error, count)

	for _, s := range services {
		go func(s Service) {
			<-s.Done()
			errors <- s.Error()
		}(s)
	}

	for err := range errors {
		if err != nil {
			return err
		}

		count--
		if count == 0 {
			close(errors)
		}
	}

	return nil
}
