package traceutil

import "github.com/opentracing/opentracing-go/log"

const (
	subsystemKey   = "subsystem"
	messageIDKey   = "message_id"
	namespaceKey   = "namespace"
	attributesKey  = "attributes"
	payloadSizeKey = "payload.size"
	payloadDataKey = "payload.data" // TODO: log full payload with "debug tracing"
)

var (
	errorEvent = log.String("event", "error")
)
