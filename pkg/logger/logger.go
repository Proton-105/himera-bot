package logger

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/himera-bot/trading-bot/pkg/config"
	slogsentry "github.com/samber/slog-sentry/v2"
	"gopkg.in/natefinch/lumberjack.v2"
)

// New constructs a slog.Logger configured according to the provided application configuration.
func New(cfg config.Config) *slog.Logger {
	level := resolveLevel(cfg.Logger.Level)

	logWriter := resolveWriter(cfg.Logger.Format)

	handlerOpts := &slog.HandlerOptions{Level: level}

	var baseHandler slog.Handler
	if cfg.Logger.Format == "json" {
		baseHandler = slog.NewJSONHandler(logWriter, handlerOpts)
	} else {
		baseHandler = slog.NewTextHandler(logWriter, handlerOpts)
	}

	maskingHandler := NewMaskingHandler(baseHandler)

	handlers := []slog.Handler{maskingHandler}

	if cfg.Sentry.Enabled {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:         cfg.Sentry.DSN,
			Environment: cfg.AppEnv,
		}); err != nil {
			slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})).Error(
				"failed to initialize sentry",
				slog.String("error", err.Error()),
			)
		} else {
			handlers = append(handlers, slogsentry.Option{Level: slog.LevelError}.NewSentryHandler())
		}
	}

	var handler slog.Handler
	if len(handlers) == 1 {
		handler = handlers[0]
	} else {
		handler = newMultiHandler(handlers...)
	}

	return slog.New(handler)
}

func resolveLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func resolveWriter(format string) io.Writer {
	if format == "json" {
		return &lumberjack.Logger{
			Filename:   "storage/logs/chimera.log",
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		}
	}

	return os.Stdout
}

type multiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(handlers ...slog.Handler) slog.Handler {
	return &multiHandler{handlers: handlers}
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}

	return false
}

func (h *multiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, record.Clone()); err != nil {
			return err
		}
	}

	return nil
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nextHandlers := make([]slog.Handler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		nextHandlers = append(nextHandlers, handler.WithAttrs(attrs))
	}

	return &multiHandler{handlers: nextHandlers}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	nextHandlers := make([]slog.Handler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		nextHandlers = append(nextHandlers, handler.WithGroup(name))
	}

	return &multiHandler{handlers: nextHandlers}
}
