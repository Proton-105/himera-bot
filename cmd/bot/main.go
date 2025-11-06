// cmd/bot/main.go

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/getsentry/sentry-go"
	"github.com/himera-bot/trading-bot/internal/middleware"
	"github.com/himera-bot/trading-bot/internal/ratelimit"
	"github.com/himera-bot/trading-bot/internal/repository"
	"github.com/himera-bot/trading-bot/pkg/config"
	"github.com/himera-bot/trading-bot/pkg/logger"
	redisclient "github.com/himera-bot/trading-bot/pkg/redis"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	bootstrapCfg := config.Config{
		AppEnv: "bootstrap",
		Logger: config.LoggerConfig{Level: "info", Format: "text"},
		Sentry: config.SentryConfig{Enabled: false},
	}
	bootstrapLog := logger.New(bootstrapCfg).With(slog.String("component", "bootstrap"))

	bootstrapLog.Info("loading configuration")

	cfg, v, err := config.Load()
	if err != nil {
		bootstrapLog.Error("failed to load config", "error", err)
		return
	}

	log := logger.New(*cfg).With(slog.String("component", "bootstrap"))
	log.Info("starting chimera bot", slog.String("config", cfg.String()))
	log.Info("configuration loaded", slog.String("env", cfg.AppEnv))

	loggingMiddleware := middleware.New(log)

	v.WatchConfig()
	configLog := log.With(slog.String("subsystem", "config_watcher"))
	v.OnConfigChange(func(event fsnotify.Event) {
		configLog.Info("configuration change detected", slog.String("event", event.String()))
		if reloadErr := v.Unmarshal(&cfg); reloadErr != nil {
			configLog.Error("failed to reload config", "error", reloadErr)
			return
		}
		configLog.Info("configuration reloaded", slog.String("config", cfg.String()))
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := sql.Open("postgres", cfg.Database.DSN())
	if err != nil {
		log.Error("failed to open database", "error", err)
		return
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			log.Error("error closing database connection", "error", cerr)
		}
	}()

	if err = db.PingContext(ctx); err != nil {
		log.Error("failed to ping database", "error", err)
		return
	}
	log.Info("connected to database",
		slog.String("host", cfg.Database.Host),
		slog.String("database", cfg.Database.Name),
	)

	coreRedisClient, err := redisclient.New(ctx, cfg.Redis.ToClientConfig())
	if err != nil {
		log.Error("failed to connect to redis", "error", err)
		return
	}
	redisClient := redisclient.NewMetricsClient(coreRedisClient)
	defer func() {
		if cerr := redisClient.Close(); cerr != nil {
			log.Error("error closing redis client", "error", cerr)
		}
	}()
	log.Info("connected to redis",
		slog.String("host", cfg.Redis.Host),
		slog.Int("db", cfg.Redis.DB),
	)

	_ = repository.NewStateRepository(coreRedisClient)
	log.Info("state repository initialized")

	_ = ratelimit.NewLimiter(coreRedisClient, 10, time.Second)
	log.Info("rate limiter initialized", slog.Int("limit", 10), slog.Duration("window", time.Second))

	log.Info("performing test redis operations for metrics")
	if err := redisClient.Set(ctx, "test_key", "test_value", 10*time.Second); err != nil {
		log.Error("redis set error", "error", err, slog.String("key", "test_key"))
	}
	if _, err := redisClient.Get(ctx, "test_key"); err != nil {
		log.Error("redis get error", "error", err, slog.String("key", "test_key"))
	}
	if _, err := redisClient.Get(ctx, "non_existent_key"); err != nil {
		log.Warn("redis get miss", "error", err, slog.String("key", "non_existent_key"))
	}
	if err := redisClient.Delete(ctx, "test_key"); err != nil {
		log.Error("redis delete error", "error", err, slog.String("key", "test_key"))
	}

	metricsLog := log.With(slog.String("subsystem", "metrics_http"))

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", logger.Middleware(loggingMiddleware(promhttp.Handler())))

		server := &http.Server{
			Addr:    fmt.Sprintf(":%s", cfg.Server.MetricsPort),
			Handler: mux,
		}

		go func() {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
				metricsLog.Error("metrics server shutdown error", "error", err)
			}
		}()

		metricsLog.Info("metrics server listening", slog.String("port", cfg.Server.MetricsPort))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			metricsLog.Error("metrics server error", "error", err)
		}
	}()

	log.Info("application started", slog.String("metrics_port", cfg.Server.MetricsPort))

	<-ctx.Done()

	sentry.Flush(2 * time.Second)

	log.Info("shutdown signal received", "error", ctx.Err())
}
