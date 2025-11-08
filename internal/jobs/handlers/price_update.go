package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"

	"github.com/Proton-105/himera-bot/internal/jobs"
)

type PriceUpdateHandler struct {
	log *slog.Logger
	// attach other dependencies here (API clients, services, etc.) when needed
}

func NewPriceUpdateHandler(log *slog.Logger) *PriceUpdateHandler {
	return &PriceUpdateHandler{log: log}
}

func (h *PriceUpdateHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload jobs.PriceUpdatePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		if h.log != nil {
			h.log.ErrorContext(ctx, "price update: failed to decode payload", slog.Any("task_type", t.Type()), slog.String("error", err.Error()))
		}
		return err
	}

	if h.log != nil {
		traceID, _ := ctx.Value("trace_id").(string)
		attrs := []slog.Attr{
			slog.String("task_type", t.Type()),
			slog.Any("addresses", payload.TokenAddresses),
			slog.Int("addresses_len", len(payload.TokenAddresses)),
		}
		if traceID != "" {
			attrs = append(attrs, slog.String("trace_id", traceID))
		}
		args := make([]any, 0, len(attrs))
		for _, attr := range attrs {
			args = append(args, attr)
		}
		h.log.InfoContext(ctx, "updating prices", args...)
	}

	time.Sleep(1 * time.Second)

	return nil
}
