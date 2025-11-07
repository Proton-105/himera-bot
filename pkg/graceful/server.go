package graceful

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// Server wraps http.Server with graceful shutdown capabilities.
type Server struct {
	httpServer      *http.Server
	log             *slog.Logger
	shutdownTimeout time.Duration
}

// NewServer constructs a graceful server wrapper.
func NewServer(log *slog.Logger, srv *http.Server, shutdownTimeout time.Duration) *Server {
	if log == nil {
		log = slog.Default()
	}

	return &Server{
		httpServer:      srv,
		log:             log,
		shutdownTimeout: shutdownTimeout,
	}
}

// ListenAndServe starts the HTTP server and handles graceful shutdown when ctx is canceled.
func (s *Server) ListenAndServe(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, 1)
	var once sync.Once

	go func() {
		if s.log != nil {
			s.log.Info("http server listening", slog.String("addr", s.httpServer.Addr))
		}

		err := s.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			if s.log != nil {
				s.log.Error("http server error", slog.Any("error", err))
			}
		}

		once.Do(func() { errCh <- err })
	}()

	<-ctx.Done()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancelShutdown()

	if s.log != nil {
		s.log.Info("shutting down http server", slog.Duration("timeout", s.shutdownTimeout))
	}

	shutdownErr := s.httpServer.Shutdown(shutdownCtx)
	if shutdownErr != nil && s.log != nil {
		s.log.Error("http server shutdown error", slog.Any("error", shutdownErr))
	}

	var listenErr error
	select {
	case listenErr = <-errCh:
	default:
	}

	if shutdownErr != nil {
		return shutdownErr
	}

	return listenErr
}
