package middleware

import (
	"time"

	telebot "gopkg.in/telebot.v3"

	"github.com/Proton-105/himera-bot/internal/bot/handlers"
	"github.com/Proton-105/himera-bot/pkg/metrics"
)

// Metrics measures execution time and status for bot handlers, reporting them to Prometheus.
func Metrics(next handlers.Handler) handlers.Handler {
	if next == nil {
		return nil
	}

	return func(c telebot.Context) error {
		start := time.Now()
		err := next(c)

		command := extractCommandName(c)
		status := "ok"
		if err != nil {
			status = "error"
		}

		metrics.RecordCommand(command, status, time.Since(start))

		return err
	}
}

func extractCommandName(c telebot.Context) string {
	if c == nil {
		return "unknown"
	}

	if cb := c.Callback(); cb != nil && cb.Data != "" {
		return cb.Data
	}

	if text := c.Text(); text != "" {
		return text
	}

	return "unknown"
}
