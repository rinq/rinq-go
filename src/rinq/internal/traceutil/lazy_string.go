package traceutil

import "github.com/opentracing/opentracing-go/log"

func lazyString(key string, s func() string) log.Field {
	return log.Lazy(func(e log.Encoder) {
		e.EmitString(key, s())
	})
}
