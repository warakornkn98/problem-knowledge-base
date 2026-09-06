// Package database owns the PostgreSQL connection and the SQL migration runner.
package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/team/pkb/internal/config"
)

// Open establishes a GORM connection pool against PostgreSQL.
func Open(cfg config.DBConfig, appEnv string) (*gorm.DB, error) {
	logLevel := gormlogger.Warn
	if appEnv == "development" {
		logLevel = gormlogger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger:                 gormlogger.Default.LogMode(logLevel),
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return db, nil
}

// WaitForReady retries Ping until the database accepts connections or the
// attempt budget is exhausted. Useful when the app boots alongside Postgres.
func WaitForReady(cfg config.DBConfig, attempts int, wait time.Duration) (*gorm.DB, error) {
	var lastErr error
	for i := 0; i < attempts; i++ {
		db, err := Open(cfg, "production")
		if err == nil {
			return db, nil
		}
		lastErr = err
		time.Sleep(wait)
	}
	return nil, fmt.Errorf("database not ready after %d attempts: %w", attempts, lastErr)
}
