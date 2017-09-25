package traceutil

import (
	"fmt"

	"github.com/opentracing/opentracing-go/log"
)

func lazyString(key string, val fmt.Stringer) log.Field {
	return log.Lazy(func(e log.Encoder) {
		e.EmitString(key, val.String())
	})
}
