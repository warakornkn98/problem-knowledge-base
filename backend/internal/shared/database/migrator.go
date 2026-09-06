package database

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"gorm.io/gorm"
)

// Migration is a single forward migration parsed from an embedded .sql file.
type Migration struct {
	Version string // e.g. "0001_init"
	SQL     string
}

// LoadMigrations reads every *.sql file from fsys and returns them ordered by
// filename (which encodes the sequence, e.g. 0001_, 0002_).
func LoadMigrations(fsys fs.FS) ([]Migration, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	var migs []Migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		body, err := fs.ReadFile(fsys, e.Name())
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		migs = append(migs, Migration{
			Version: strings.TrimSuffix(e.Name(), ".sql"),
			SQL:     string(body),
		})
	}
	sort.Slice(migs, func(i, j int) bool { return migs[i].Version < migs[j].Version })
	return migs, nil
}

// Migrate applies every not-yet-applied migration inside its own transaction and
// records it in schema_migrations. Returns the versions applied by this call.
//
// It runs against the raw *sql.DB (not GORM) so a migration file may contain
// multiple statements — pgx uses the simple query protocol when there are no
// bind parameters.
func Migrate(db *gorm.DB, fsys fs.FS) ([]string, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	ctx := context.Background()

	if _, err := sqlDB.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`); err != nil {
		return nil, fmt.Errorf("ensure schema_migrations: %w", err)
	}

	rows, err := sqlDB.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("load applied migrations: %w", err)
	}
	applied := map[string]bool{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return nil, err
		}
		applied[v] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	migs, err := LoadMigrations(fsys)
	if err != nil {
		return nil, err
	}

	var justApplied []string
	for _, m := range migs {
		if applied[m.Version] {
			continue
		}

		tx, err := sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return justApplied, err
		}
		if _, err := tx.ExecContext(ctx, m.SQL); err != nil {
			_ = tx.Rollback()
			return justApplied, fmt.Errorf("apply %s: %w", m.Version, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, m.Version); err != nil {
			_ = tx.Rollback()
			return justApplied, fmt.Errorf("record %s: %w", m.Version, err)
		}
		if err := tx.Commit(); err != nil {
			return justApplied, fmt.Errorf("commit %s: %w", m.Version, err)
		}
		justApplied = append(justApplied, m.Version)
	}
	return justApplied, nil
}
