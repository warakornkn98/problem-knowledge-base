// Package domain defines the Tag entity and its persistence port.
package domain

import (
	"context"
	"errors"
	"time"
)

// Tag is a free-form label attached to problems.
type Tag struct {
	ID        int64
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time

	// UsageCount is populated only by list queries that ask for it.
	UsageCount int64
}

// Domain errors.
var (
	ErrNotFound  = errors.New("tag not found")
	ErrNameTaken = errors.New("tag name already exists")
)

// Repository is the persistence port for tags.
type Repository interface {
	Create(ctx context.Context, t *Tag) error
	Update(ctx context.Context, t *Tag) error
	Delete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*Tag, error)
	FindByName(ctx context.Context, name string) (*Tag, error)
	List(ctx context.Context, query string, withCounts bool) ([]Tag, error)

	// EnsureByNames returns the tags for the given names, creating any that do
	// not exist yet. Names are normalised (trimmed, lower-cased) by the caller.
	EnsureByNames(ctx context.Context, names []string) ([]Tag, error)
	FindByIDs(ctx context.Context, ids []int64) ([]Tag, error)
}
