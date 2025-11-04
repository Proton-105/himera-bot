package logger

import (
	"log"
	"os"
)

// Logger wraps the standard log.Logger with Himera-specific defaults.
type Logger struct {
	*log.Logger
}

// New creates a Logger that writes to stdout with the Himera prefix and detailed flags.
func New() *Logger {
	base := log.New(os.Stdout, "[himera] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

	return &Logger{Logger: base}
}
