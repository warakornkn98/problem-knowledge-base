// Command migrate applies the embedded SQL migrations.
//
//	go run ./cmd/migrate up     (default)
//	go run ./cmd/migrate status
package main

import (
	"fmt"
	"os"

	"github.com/team/pkb/internal/config"
	"github.com/team/pkb/internal/shared/database"
	"github.com/team/pkb/internal/shared/logger"
	"github.com/team/pkb/migrations"
)

func main() {
	cmd := "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

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

	switch cmd {
	case "up":
		applied, err := database.Migrate(db, migrations.FS)
		if err != nil {
			log.Error("migrate failed", "error", err)
			os.Exit(1)
		}
		if len(applied) == 0 {
			fmt.Println("already up to date")
			return
		}
		fmt.Printf("applied %d migration(s):\n", len(applied))
		for _, v := range applied {
			fmt.Println("  +", v)
		}
	case "status":
		migs, err := database.LoadMigrations(migrations.FS)
		if err != nil {
			log.Error("load migrations failed", "error", err)
			os.Exit(1)
		}
		var applied []string
		_ = db.Raw(`SELECT version FROM schema_migrations ORDER BY version`).Scan(&applied).Error
		done := map[string]bool{}
		for _, v := range applied {
			done[v] = true
		}
		for _, m := range migs {
			mark := " "
			if done[m.Version] {
				mark = "x"
			}
			fmt.Printf("[%s] %s\n", mark, m.Version)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q (use: up | status)\n", cmd)
		os.Exit(2)
	}
}
