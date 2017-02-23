package deferutil

import "sync"

// Lock acquires a lock and returns a function that releases the lock the first
// time it is called.
//
// This can be be used to offer panic-safe mutex locking, but also unlock the
// mutex before the end of the function if necessary.
func Lock(l sync.Locker) func() {
	l.Lock()

	return func() {
		if l != nil {
			l.Unlock()
			l = nil
		}
	}
}

// RLock is a variant of Lock that operates on a read-write mutex's read locker.
func RLock(m *sync.RWMutex) func() {
	return Lock(m.RLocker())
}
