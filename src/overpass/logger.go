package overpass

import (
	"log"
	"os"
)

// Logger is a simple interface used for logging throughout Overpass.
type Logger interface {
	Log(f string, v ...interface{})
	IsDebug() bool
}

// NewLogger returns a logger that writes to stdout.
func NewLogger(isDebug bool) Logger {
	return standardLogger{
		isDebug,
		log.New(os.Stdout, "", log.LstdFlags),
	}
}

type standardLogger struct {
	isDebug bool
	logger  *log.Logger
}

func (l standardLogger) Log(f string, v ...interface{}) {
	l.logger.Printf(f, v...)
}

func (l standardLogger) IsDebug() bool {
	return l.isDebug
}
