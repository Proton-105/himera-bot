package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Proton-105/himera-bot/internal/bot"
	"github.com/Proton-105/himera-bot/internal/health"
	"github.com/Proton-105/himera-bot/internal/lifecycle"
	"github.com/Proton-105/himera-bot/internal/middleware"
	"github.com/Proton-105/himera-bot/internal/ratelimit"
	"github.com/Proton-105/himera-bot/internal/state"
	"github.com/Proton-105/himera-bot/pkg/config"
	"github.com/Proton-105/himera-bot/pkg/logger"
	redisclient "github.com/Proton-105/himera-bot/pkg/redis"
	"github.com/fsnotify/fsnotify"
	"github.com/getsentry/sentry-go"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	code := run()
	os.Exit(code)
}

func run() int {
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
		return 0
	}

	log := logger.New(*cfg).With(slog.String("component", "bootstrap"))
	log.Info("starting chimera bot", slog.String("config", cfg.String()))
	log.Info("configuration loaded", slog.String("env", cfg.AppEnv))

	loggingMiddleware := middleware.New(log)
	shutdownCoordinator := lifecycle.NewShutdown(log.With(slog.String("component", "shutdown")))

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
		return 0
	}
	shutdownCoordinator.Register("db-close", func(ctx context.Context) error {
		if db == nil {
			return nil
		}
		return db.Close()
	})

	if err = db.PingContext(ctx); err != nil {
		log.Error("failed to ping database", "error", err)
		return 0
	}
	log.Info("connected to database",
		slog.String("host", cfg.Database.Host),
		slog.String("database", cfg.Database.Name),
	)

	coreRedisClient, err := redisclient.New(ctx, cfg.Redis.ToClientConfig())
	if err != nil {
		log.Error("failed to connect to redis", "error", err)
		return 0
	}
	redisClient := redisclient.NewMetricsClient(coreRedisClient)
	shutdownCoordinator.Register("redis-close", func(ctx context.Context) error {
		if redisClient == nil {
			return nil
		}
		return redisClient.Close()
	})
	log.Info("connected to redis",
		slog.String("host", cfg.Redis.Host),
		slog.Int("db", cfg.Redis.DB),
	)

	stateStorage := state.NewRedisStorage(coreRedisClient.Raw(), log)
	fsm := state.NewStateMachine(stateStorage, log, coreRedisClient.Raw())
	log.Info("state machine initialized")

	cleaner := state.NewCleaner(coreRedisClient.Raw(), stateStorage, log, time.Hour, 5*time.Minute)
	go cleaner.Run(ctx)
	log.Info("state cleaner started", slog.Duration("ttl", time.Hour), slog.Duration("interval", 5*time.Minute))

	rules := ratelimit.NewRules(cfg.RateLimit)
	redisLimiter := ratelimit.NewRedisLimiter(coreRedisClient.Raw(), log)
	memoryLimiter := ratelimit.NewMemoryLimiter(log)
	adaptiveLimiter := ratelimit.NewAdaptiveLimiter(redisLimiter, memoryLimiter, log)
	rateLimitMw := middleware.NewRateLimitMiddleware(adaptiveLimiter, rules, log)
	rateLimitCleaner := ratelimit.NewCleaner(coreRedisClient.Raw(), log, time.Minute)
	go rateLimitCleaner.Run(ctx)
	log.Info("rate limit cleaner started", slog.Duration("interval", time.Minute))

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

	tgBot, err := bot.New(*cfg, log, db, fsm, rateLimitMw)
	if err != nil {
		log.Error("failed to create telegram bot", "error", err)
		return 0
	}

	checker := health.NewChecker(log)
	checker.AddCheck("database", health.NewDBChecker(db))
	checker.AddCheck("redis", health.NewRedisChecker(coreRedisClient))
	checker.AddCheck("telegram", health.NewTelegramChecker(tgBot.Telebot()))

	go tgBot.Start()
	log.Info("telegram bot started")

	metricsLog := log.With(slog.String("subsystem", "metrics_http"))

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", logger.Middleware(loggingMiddleware(promhttp.Handler())))
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			results := checker.Check(r.Context())

			status := http.StatusOK
			for _, result := range results {
				if result != "OK" {
					status = http.StatusServiceUnavailable
					break
				}
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)

			if err := json.NewEncoder(w).Encode(results); err != nil {
				metricsLog.Error("failed to write health response", slog.Any("error", err))
			}
		})

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

	tgBot.Stop()
	sentry.Flush(2 * time.Second)

	log.Info("shutdown signal received", "error", ctx.Err())

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	shutdownStart := time.Now()
	if err := shutdownCoordinator.Execute(shutdownCtx); err != nil {
		log.Error("shutdown completed with errors", slog.Any("error", err), slog.Duration("elapsed", time.Since(shutdownStart)))
		cancel()
		return 1
	}

	log.Info("shutdown completed successfully", slog.Duration("elapsed", time.Since(shutdownStart)))
	cancel()
	return 0
}
