// Command api is the HTTP entrypoint for the Problem Knowledge Base backend.
package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/team/pkb/internal/config"
	"github.com/team/pkb/internal/seed"
	"github.com/team/pkb/internal/server"
	"github.com/team/pkb/internal/shared/database"
	"github.com/team/pkb/internal/shared/logger"
	"github.com/team/pkb/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log := logger.New(cfg.AppEnv)

	db, err := database.WaitForReady(cfg.DB, 30, 2*time.Second)
	if err != nil {
		log.Error("database connection failed", "error", err)
		os.Exit(1)
	}

	if cfg.AutoMigrate {
		applied, err := database.Migrate(db, migrations.FS)
		if err != nil {
			log.Error("auto-migrate failed", "error", err)
			os.Exit(1)
		}
		if len(applied) > 0 {
			log.Info("migrations applied", "versions", applied)
		}
	}

	if cfg.AutoSeed {
		if err := seed.Run(context.Background(), db, cfg); err != nil {
			log.Error("auto-seed failed", "error", err)
			os.Exit(1)
		}
	}

	app := server.New(cfg, db, log)

	go func() {
		addr := ":" + cfg.AppPort
		log.Info("listening", "addr", addr)
		if err := app.Listen(addr); err != nil && !errors.Is(err, context.Canceled) {
			log.Error("server stopped", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
	}
}
