package logger

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// correlationIDKey marks the context storage slot for the correlation identifier.
type correlationIDKey struct{}

// CorrelationIDFromContext returns the correlation identifier stored in ctx, or an empty string when absent.
func CorrelationIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDKey{}).(string); ok {
		return id
	}

	return ""
}

// Middleware injects a correlation identifier into the request context before delegating to the next handler.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := uuid.NewString()
		ctxWithID := context.WithValue(r.Context(), correlationIDKey{}, correlationID)
		next.ServeHTTP(w, r.WithContext(ctxWithID))
	})
}
