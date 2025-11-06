package logger

import (
	"context"
	"log/slog"
	"strings"
)

var sensitiveKeys = []string{
	"password",
	"token",
	"secret",
	"api_key",
	"authorization",
}

// MaskingHandler wraps a slog.Handler and masks sensitive attributes before delegating.
type MaskingHandler struct {
	next slog.Handler
}

// NewMaskingHandler creates a handler that masks sensitive fields before passing records downstream.
func NewMaskingHandler(next slog.Handler) *MaskingHandler {
	return &MaskingHandler{next: next}
}

// Enabled reports whether the handler handles records at the given level.
func (h *MaskingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

// WithAttrs returns a new handler with additional attributes.
func (h *MaskingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &MaskingHandler{next: h.next.WithAttrs(attrs)}
}

// WithGroup returns a new handler with an appended group name.
func (h *MaskingHandler) WithGroup(name string) slog.Handler {
	return &MaskingHandler{next: h.next.WithGroup(name)}
}

// Handle applies masking to sensitive attributes and delegates to the wrapped handler.
func (h *MaskingHandler) Handle(ctx context.Context, record slog.Record) error {
	masked := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)

	record.Attrs(func(attr slog.Attr) bool {
		maskedAttr := attr
		if isSensitiveKey(attr.Key) {
			maskedAttr.Value = slog.StringValue("***")
		}
		masked.AddAttrs(maskedAttr)
		return true
	})

	return h.next.Handle(ctx, masked)
}

func isSensitiveKey(key string) bool {
	for _, sensitive := range sensitiveKeys {
		if strings.EqualFold(key, sensitive) {
			return true
		}
	}
	return false
}
