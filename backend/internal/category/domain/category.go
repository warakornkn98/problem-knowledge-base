// Package domain defines the Category entity and the ports the application
// layer depends on. Nothing here imports a framework or a database driver.
package domain

import (
	"context"
	"errors"
	"time"
)

// Category groups problems by area (Backend, Network, Docker, ...).
type Category struct {
	ID          int64
	Name        string
	Slug        string
	Description string
	Color       string
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// ProblemCount is populated only by list queries that ask for it.
	ProblemCount int64
}

// Domain errors.
var (
	ErrNotFound  = errors.New("category not found")
	ErrNameTaken = errors.New("category name already exists")
	ErrInUse     = errors.New("category still has problems attached")
)

// Repository is the persistence port for categories.
type Repository interface {
	Create(ctx context.Context, c *Category) error
	Update(ctx context.Context, c *Category) error
	Delete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*Category, error)
	FindByName(ctx context.Context, name string) (*Category, error)
	List(ctx context.Context, withCounts bool) ([]Category, error)
	Exists(ctx context.Context, id int64) (bool, error)
	CountProblems(ctx context.Context, id int64) (int64, error)
}
