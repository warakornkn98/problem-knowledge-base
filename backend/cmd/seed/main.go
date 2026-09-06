// Command seed inserts baseline data: categories, tags, a bootstrap admin
// account and one example problem. Safe to run multiple times.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/team/pkb/internal/config"
	"github.com/team/pkb/internal/seed"
	"github.com/team/pkb/internal/shared/database"
	"github.com/team/pkb/internal/shared/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log := logger.New(cfg.AppEnv)

	db, err := database.Open(cfg.DB, cfg.AppEnv)
	if err != nil {
		log.Error("connect failed", "error", err)
		os.Exit(1)
	}

	if err := seed.Run(context.Background(), db, cfg); err != nil {
		log.Error("seed failed", "error", err)
		os.Exit(1)
	}
	fmt.Println("seed complete")
	fmt.Printf("admin login: %s / %s\n", cfg.SeedAdminUsername, cfg.SeedAdminPassword)
}
