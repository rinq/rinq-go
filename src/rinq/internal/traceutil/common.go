package traceutil

import "github.com/opentracing/opentracing-go/log"

var (
	successEvent = log.String("event", "success")
	errorEvent   = log.String("event", "error")
)
