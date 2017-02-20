package overpass

import "log"

// Logger is a simple interface used for logging throughout Overpass.
type Logger interface {
	Log(f string, v ...interface{})
	IsDebug() bool
}

// NewLogger returns a logger that writes using the built-in logger.
func NewLogger(isDebug bool) Logger {
	return standardLogger{isDebug}
}

type standardLogger struct {
	isDebug bool
}

func (l standardLogger) Log(f string, v ...interface{}) {
	log.Printf(f, v...)
}

func (l standardLogger) IsDebug() bool {
	return l.isDebug
}
