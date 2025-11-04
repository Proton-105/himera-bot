package main

import (
	"context"
	"database/sql"
	"os/signal"
	"syscall"

	"github.com/himera-bot/trading-bot/internal/database"
	"github.com/himera-bot/trading-bot/pkg/config"
	"github.com/himera-bot/trading-bot/pkg/logger"

	_ "github.com/lib/pq"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	log := logger.New()

	log.Printf("Starting Himera trading bot (env=%s, http_port=%s, log_level=%s)", cfg.Env, cfg.HTTPPort, cfg.LogLevel)

	db, err := sql.Open("postgres", cfg.GetDBConnectionString())
	if err != nil {
		log.Printf("failed to open database: %v", err)
		return
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			log.Printf("error closing database: %v", cerr)
		}
	}()

	if err := db.PingContext(ctx); err != nil {
		log.Printf("failed to ping database: %v", err)
		return
	}

	migrator := database.NewMigrator(db, log)
	if err := migrator.ApplyDir(ctx, "migrations"); err != nil {
		log.Printf("failed to apply migrations: %v", err)
		return
	}
	log.Println("Database migrations applied successfully.")

	<-ctx.Done()

	log.Println("Himera trading bot shutting down.")
}
