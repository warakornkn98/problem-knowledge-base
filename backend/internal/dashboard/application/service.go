// Package application implements the dashboard read models.
package application

import "context"

// Summary is the headline counter block.
type Summary struct {
	Total         int64 `json:"total"`
	Open          int64 `json:"open"`
	Investigating int64 `json:"investigating"`
	Solved        int64 `json:"solved"`
	Known         int64 `json:"known"`
	Today         int64 `json:"today"`
	ThisWeek      int64 `json:"this_week"`
	Unsolved      int64 `json:"unsolved"`
}

// Bucket is a labelled count (category / project chart rows).
type Bucket struct {
	Label string `json:"label"`
	Slug  string `json:"slug,omitempty"`
	Count int64  `json:"count"`
}

// ErrorBucket is a recurring error message with its occurrence count.
type ErrorBucket struct {
	ErrorMessage string `json:"error_message"`
	Count        int64  `json:"count"`
}

// Repository is the read port for dashboard aggregates.
type Repository interface {
	Summary(ctx context.Context) (Summary, error)
	TopCategories(ctx context.Context, limit int) ([]Bucket, error)
	TopProjects(ctx context.Context, limit int) ([]Bucket, error)
	CommonErrors(ctx context.Context, limit int) ([]ErrorBucket, error)
}

// Service exposes the dashboard queries.
type Service struct {
	repo Repository
}

// NewService wires the dashboard service.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// Summary returns the counter block.
func (s *Service) Summary(ctx context.Context) (Summary, error) {
	return s.repo.Summary(ctx)
}

// TopCategories returns the most problematic categories.
func (s *Service) TopCategories(ctx context.Context, limit int) ([]Bucket, error) {
	return s.repo.TopCategories(ctx, clamp(limit, 5, 20))
}

// TopProjects returns the most problematic projects.
func (s *Service) TopProjects(ctx context.Context, limit int) ([]Bucket, error) {
	return s.repo.TopProjects(ctx, clamp(limit, 5, 20))
}

// CommonErrors returns the most frequently recorded error messages.
func (s *Service) CommonErrors(ctx context.Context, limit int) ([]ErrorBucket, error) {
	return s.repo.CommonErrors(ctx, clamp(limit, 5, 20))
}

func clamp(v, def, max int) int {
	if v <= 0 {
		return def
	}
	if v > max {
		return max
	}
	return v
}
