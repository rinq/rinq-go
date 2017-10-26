package service

import "errors"

// ErrStopped is returned by any operation that can not be fulfilled because
// the service that provides it is stopping or has already stopped.
var ErrStopped = errors.New("service has been stopped")
