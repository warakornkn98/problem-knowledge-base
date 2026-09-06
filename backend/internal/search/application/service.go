// Package application implements the unified search use case. Search is the
// most important capability of the knowledge base, so it gets its own feature
// slice even though it reads through the problem repository port.
package application

import (
	"context"

	problemdomain "github.com/team/pkb/internal/problem/domain"
)

// Service runs a query and returns matches together with facet counts for the
// advanced-search sidebar.
type Service struct {
	problems problemdomain.Repository
}

// NewService wires the search service.
func NewService(problems problemdomain.Repository) *Service {
	return &Service{problems: problems}
}

// Result bundles a page of matches with the query's facet aggregation.
type Result struct {
	Items  []problemdomain.ListItem
	Total  int64
	Facets problemdomain.Facets
}

// Search executes q, returning the requested page plus facets computed over the
// full match set.
func (s *Service) Search(ctx context.Context, q problemdomain.ListQuery) (Result, error) {
	list, err := s.problems.List(ctx, q)
	if err != nil {
		return Result{}, err
	}
	facets, err := s.problems.Facets(ctx, q)
	if err != nil {
		return Result{}, err
	}
	return Result{Items: list.Items, Total: list.Total, Facets: facets}, nil
}
