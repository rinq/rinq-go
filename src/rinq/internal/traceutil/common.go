package traceutil

import "github.com/opentracing/opentracing-go/log"

const (
	subsystemKey   = "subsystem"
	messageIDKey   = "message_id"
	namespaceKey   = "namespace"
	payloadSizeKey = "payload.size"
)

var (
	errorEvent = log.String("event", "error")
)
