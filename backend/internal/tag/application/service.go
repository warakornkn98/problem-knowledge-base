// Package application holds the tag use cases.
package application

import (
	"context"
	"errors"
	"strings"

	"github.com/team/pkb/internal/shared/httpx"
	"github.com/team/pkb/internal/shared/slug"
	"github.com/team/pkb/internal/tag/domain"
)

// Service exposes tag operations.
type Service struct {
	repo domain.Repository
}

// NewService wires the tag service.
func NewService(repo domain.Repository) *Service { return &Service{repo: repo} }

// List returns tags filtered by an optional name fragment.
func (s *Service) List(ctx context.Context, query string, withCounts bool) ([]domain.Tag, error) {
	return s.repo.List(ctx, strings.TrimSpace(query), withCounts)
}

// Create adds a new tag.
func (s *Service) Create(ctx context.Context, name string) (*domain.Tag, error) {
	name = normalize(name)
	if name == "" {
		return nil, httpx.NewUnprocessable("name is required")
	}
	if existing, _ := s.repo.FindByName(ctx, name); existing != nil {
		return nil, httpx.NewConflict("tag name already exists")
	}
	t := &domain.Tag{Name: name, Slug: slug.Make(name)}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, mapErr(err)
	}
	return t, nil
}

// Update renames a tag.
func (s *Service) Update(ctx context.Context, id int64, name string) (*domain.Tag, error) {
	t, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	name = normalize(name)
	if name == "" {
		return nil, httpx.NewUnprocessable("name is required")
	}
	if other, _ := s.repo.FindByName(ctx, name); other != nil && other.ID != id {
		return nil, httpx.NewConflict("tag name already exists")
	}
	t.Name = name
	t.Slug = slug.Make(name)
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, mapErr(err)
	}
	return t, nil
}

// Delete removes a tag (its problem associations cascade away).
func (s *Service) Delete(ctx context.Context, id int64) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return mapErr(err)
	}
	return mapErr(s.repo.Delete(ctx, id))
}

// Resolve turns a mix of tag ids and free-form names into a de-duplicated tag
// set, creating missing names on the fly. Used by the problem module.
func (s *Service) Resolve(ctx context.Context, ids []int64, names []string) ([]domain.Tag, error) {
	result := make(map[int64]domain.Tag)

	if len(ids) > 0 {
		found, err := s.repo.FindByIDs(ctx, ids)
		if err != nil {
			return nil, mapErr(err)
		}
		for _, t := range found {
			result[t.ID] = t
		}
	}

	cleanNames := make([]string, 0, len(names))
	seen := map[string]bool{}
	for _, n := range names {
		n = normalize(n)
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		cleanNames = append(cleanNames, n)
	}
	if len(cleanNames) > 0 {
		ensured, err := s.repo.EnsureByNames(ctx, cleanNames)
		if err != nil {
			return nil, mapErr(err)
		}
		for _, t := range ensured {
			result[t.ID] = t
		}
	}

	out := make([]domain.Tag, 0, len(result))
	for _, t := range result {
		out = append(out, t)
	}
	return out, nil
}

func normalize(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func mapErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrNotFound):
		return httpx.NewNotFound("tag not found")
	case errors.Is(err, domain.ErrNameTaken):
		return httpx.NewConflict("tag name already exists")
	default:
		return httpx.NewInternal(err)
	}
}
