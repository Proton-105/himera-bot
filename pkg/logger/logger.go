package logger

import (
	"log/slog"
	"os"
)

// Logger wraps slog.Logger providing Himera-specific defaults and helpers.
type Logger struct {
	*slog.Logger
}

type options struct {
	level     slog.Leveler
	handler   slog.Handler
	addSource bool
}

// Option configures logger creation.
type Option func(*options)

// WithLevel overrides the default logging level.
func WithLevel(level slog.Leveler) Option {
	return func(o *options) {
		o.level = level
	}
}

// WithHandler allows supplying a fully custom slog handler.
func WithHandler(handler slog.Handler) Option {
	return func(o *options) {
		o.handler = handler
	}
}

// WithSource toggles source reporting on log lines.
func WithSource(addSource bool) Option {
	return func(o *options) {
		o.addSource = addSource
	}
}

// New constructs a structured logger with JSON output, info level, and source data by default.
func New(opts ...Option) *Logger {
	cfg := options{
		level:     slog.LevelInfo,
		addSource: true,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	handler := cfg.handler
	if handler == nil {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     cfg.level,
			AddSource: cfg.addSource,
		})
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// With adds structured context to the logger and returns a derived instance.
func (l *Logger) With(args ...any) *Logger {
	if l == nil {
		return nil
	}

	return &Logger{Logger: l.Logger.With(args...)}
}

// WithGroup nests subsequent attributes under the provided group name.
func (l *Logger) WithGroup(name string) *Logger {
	if l == nil {
		return nil
	}

	return &Logger{Logger: l.Logger.WithGroup(name)}
}

// Fatal logs the message at error level and terminates the process.
func (l *Logger) Fatal(msg string, args ...any) {
	if l == nil {
		os.Exit(1)
		return
	}

	l.Logger.Error(msg, args...)
	os.Exit(1)
}

// Err builds a consistent error attribute for structured logging.
func Err(err error) slog.Attr {
	return slog.Any("error", err)
}
