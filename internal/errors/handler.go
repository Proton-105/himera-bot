package errors

import (
	"context"
	"errors"
	"log/slog"

	"github.com/getsentry/sentry-go"

	"github.com/Proton-105/himera-bot/pkg/logger"
)

type Handler struct {
	log           *slog.Logger
	sentryEnabled bool
}

func NewHandler(log *slog.Logger, sentryEnabled bool) *Handler {
	return &Handler{
		log:           log,
		sentryEnabled: sentryEnabled,
	}
}

func (h *Handler) Handle(ctx context.Context, err error) (string, bool) {
	if err == nil {
		return "", false
	}

	if ctx == nil {
		ctx = context.Background()
	}

	log := h.log
	if log == nil {
		log = slog.Default()
	}

	var appErr *AppError
	if errors.As(err, &appErr) && appErr != nil {
		attrs := []slog.Attr{
			slog.String("code", appErr.Code),
			slog.String("message", appErr.Message),
			slog.String("severity", string(appErr.Severity)),
			slog.Bool("retryable", appErr.Retryable),
		}

		if correlationID := logger.CorrelationIDFromContext(ctx); correlationID != "" {
			attrs = append(attrs, slog.String("correlation_id", correlationID))
		}

		log.Error("application error", attrsToArgs(attrs)...)

		if h.sentryEnabled && (appErr.Severity == SeverityCritical || appErr.Severity == SeverityHigh) {
			h.sendToSentry(err)
		}

		userMessage := appErr.UserMessage
		if userMessage == "" {
			userMessage = "Произошла ошибка. Попробуйте позже"
		}

		return userMessage, appErr.Retryable
	}

	attrs := []slog.Attr{
		slog.String("message", err.Error()),
		slog.String("severity", string(SeverityHigh)),
		slog.Bool("retryable", false),
	}

	if correlationID := logger.CorrelationIDFromContext(ctx); correlationID != "" {
		attrs = append(attrs, slog.String("correlation_id", correlationID))
	}

	log.Error("unknown error", attrsToArgs(attrs)...)

	if h.sentryEnabled {
		h.sendToSentry(err)
	}

	return "Произошла ошибка. Попробуйте позже", false
}

func (h *Handler) sendToSentry(err error) {
	if err == nil {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		var appErr *AppError
		if errors.As(err, &appErr) && appErr != nil {
			if appErr.Code != "" {
				scope.SetTag("code", appErr.Code)
			}

			if appErr.Severity != "" {
				scope.SetTag("severity", string(appErr.Severity))
			}
		}

		sentry.CaptureException(err)
	})
}

func attrsToArgs(attrs []slog.Attr) []any {
	args := make([]any, 0, len(attrs))
	for _, attr := range attrs {
		args = append(args, attr)
	}

	return args
}
