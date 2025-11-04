package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/himera-bot/trading-bot/pkg/config"
	"github.com/himera-bot/trading-bot/pkg/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	log := logger.New()

	log.Printf("Starting Himera trading bot (env=%s, http_port=%s, log_level=%s)", cfg.Env, cfg.HTTPPort, cfg.LogLevel)

	<-ctx.Done()

	log.Println("Himera trading bot shutting down.")
}
