// Package infrastructure is the GORM/SQL adapter for the dashboard read port.
package infrastructure

import (
	"context"

	"gorm.io/gorm"

	"github.com/team/pkb/internal/dashboard/application"
)

// Repository implements application.Repository with aggregate SQL.
type Repository struct {
	db *gorm.DB
}

// NewRepository builds the dashboard repository.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// Summary runs the counter block in a single round-trip.
func (r *Repository) Summary(ctx context.Context) (application.Summary, error) {
	var s application.Summary
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*)                                                   AS total,
			COUNT(*) FILTER (WHERE status = 'OPEN')                    AS open,
			COUNT(*) FILTER (WHERE status = 'INVESTIGATING')           AS investigating,
			COUNT(*) FILTER (WHERE status = 'SOLVED')                  AS solved,
			COUNT(*) FILTER (WHERE status = 'KNOWN')                   AS known,
			COUNT(*) FILTER (WHERE status IN ('OPEN','INVESTIGATING')) AS unsolved,
			COUNT(*) FILTER (WHERE created_at >= date_trunc('day', now()))  AS today,
			COUNT(*) FILTER (WHERE created_at >= date_trunc('week', now())) AS this_week
		FROM problems
	`).Scan(&s).Error
	return s, err
}

// TopCategories ranks categories by attached problem count.
func (r *Repository) TopCategories(ctx context.Context, limit int) ([]application.Bucket, error) {
	var out []application.Bucket
	err := r.db.WithContext(ctx).Raw(`
		SELECT c.name AS label, c.slug AS slug, COUNT(p.id) AS count
		FROM problem_categories c
		LEFT JOIN problems p ON p.category_id = c.id
		GROUP BY c.id, c.name, c.slug
		HAVING COUNT(p.id) > 0
		ORDER BY count DESC, label ASC
		LIMIT ?
	`, limit).Scan(&out).Error
	return out, err
}

// TopProjects ranks projects by problem count.
func (r *Repository) TopProjects(ctx context.Context, limit int) ([]application.Bucket, error) {
	var out []application.Bucket
	err := r.db.WithContext(ctx).Raw(`
		SELECT project AS label, COUNT(*) AS count
		FROM problems
		GROUP BY project
		ORDER BY count DESC, label ASC
		LIMIT ?
	`, limit).Scan(&out).Error
	return out, err
}

// CommonErrors groups by the first line of error_message.
func (r *Repository) CommonErrors(ctx context.Context, limit int) ([]application.ErrorBucket, error) {
	var out []application.ErrorBucket
	err := r.db.WithContext(ctx).Raw(`
		SELECT btrim(split_part(error_message, E'\n', 1)) AS error_message, COUNT(*) AS count
		FROM problems
		WHERE btrim(coalesce(error_message, '')) <> ''
		GROUP BY 1
		HAVING COUNT(*) >= 1
		ORDER BY count DESC, error_message ASC
		LIMIT ?
	`, limit).Scan(&out).Error
	return out, err
}
