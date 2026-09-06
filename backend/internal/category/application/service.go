// Package application holds the category use cases. It orchestrates the domain
// and the repository port; it does not know about HTTP.
package application

import (
	"context"
	"errors"
	"strings"

	"github.com/team/pkb/internal/category/domain"
	"github.com/team/pkb/internal/shared/httpx"
	"github.com/team/pkb/internal/shared/slug"
)

// Service exposes category operations to the transport layer.
type Service struct {
	repo domain.Repository
}

// NewService wires the category service.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// CreateInput is the payload for creating a category.
type CreateInput struct {
	Name        string
	Description string
	Color       string
}

// UpdateInput is the payload for updating a category.
type UpdateInput struct {
	Name        string
	Description string
	Color       string
}

// List returns all categories, optionally with problem counts.
func (s *Service) List(ctx context.Context, withCounts bool) ([]domain.Category, error) {
	return s.repo.List(ctx, withCounts)
}

// Get returns a single category.
func (s *Service) Get(ctx context.Context, id int64) (*domain.Category, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

// Create adds a new category.
func (s *Service) Create(ctx context.Context, in CreateInput) (*domain.Category, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, httpx.NewUnprocessable("name is required")
	}
	if existing, _ := s.repo.FindByName(ctx, name); existing != nil {
		return nil, httpx.NewConflict("category name already exists")
	}

	c := &domain.Category{
		Name:        name,
		Slug:        slug.Make(name),
		Description: strings.TrimSpace(in.Description),
		Color:       strings.TrimSpace(in.Color),
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

// Update changes an existing category.
func (s *Service) Update(ctx context.Context, id int64, in UpdateInput) (*domain.Category, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, httpx.NewUnprocessable("name is required")
	}
	if other, _ := s.repo.FindByName(ctx, name); other != nil && other.ID != id {
		return nil, httpx.NewConflict("category name already exists")
	}

	c.Name = name
	c.Slug = slug.Make(name)
	c.Description = strings.TrimSpace(in.Description)
	c.Color = strings.TrimSpace(in.Color)

	if err := s.repo.Update(ctx, c); err != nil {
		return nil, mapErr(err)
	}
	return c, nil
}

// Delete removes a category that has no problems attached.
func (s *Service) Delete(ctx context.Context, id int64) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return mapErr(err)
	}
	count, err := s.repo.CountProblems(ctx, id)
	if err != nil {
		return mapErr(err)
	}
	if count > 0 {
		return httpx.NewConflict("category still has problems attached")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return mapErr(err)
	}
	return nil
}

func mapErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrNotFound):
		return httpx.NewNotFound("category not found")
	case errors.Is(err, domain.ErrNameTaken):
		return httpx.NewConflict("category name already exists")
	case errors.Is(err, domain.ErrInUse):
		return httpx.NewConflict("category still has problems attached")
	default:
		return httpx.NewInternal(err)
	}
}
