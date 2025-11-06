package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/Proton-105/himera-bot/pkg/logger"
)

// New creates an HTTP middleware that logs request and response details.
func New(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			recorder := httptest.NewRecorder()
			next.ServeHTTP(recorder, r)

			correlationID := logger.CorrelationIDFromContext(r.Context())

			for key, values := range recorder.Header() {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}

			statusCode := recorder.Code
			if statusCode == 0 {
				statusCode = http.StatusOK
			}

			w.WriteHeader(statusCode)
			_, _ = recorder.Body.WriteTo(w)

			loggerInstance := log
			if loggerInstance == nil {
				loggerInstance = slog.Default()
			}

			loggerInstance.Info(
				"handled http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", statusCode),
				slog.Duration("duration", time.Since(start)),
				slog.String("correlation_id", correlationID),
			)
		})
	}
}
