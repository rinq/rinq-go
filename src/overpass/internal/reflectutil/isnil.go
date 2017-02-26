package reflectutil

import "reflect"

// IsNil returns true if the interface is nil or contains a value that is nil.
func IsNil(v interface{}) (isNil bool) {
	if v == nil {
		return true
	}

	defer func() {
		if r := recover(); r != nil {
			isNil = false
		}
	}()

	return reflect.ValueOf(v).IsNil()
}
