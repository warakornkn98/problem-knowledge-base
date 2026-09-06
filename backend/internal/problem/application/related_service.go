package application

import (
	"context"
	"strings"

	"github.com/team/pkb/internal/problem/domain"
	"github.com/team/pkb/internal/shared/httpx"
)

// ListRelated returns the problems explicitly linked to the given problem.
func (s *Service) ListRelated(ctx context.Context, problemID int64) ([]domain.RelatedLink, error) {
	if _, err := s.repo.FindByID(ctx, problemID); err != nil {
		return nil, mapRepoErr(err)
	}
	links, err := s.related.List(ctx, problemID)
	if err != nil {
		return nil, httpx.NewInternal(err)
	}
	return links, nil
}

// AddRelated creates a link between two problems.
func (s *Service) AddRelated(ctx context.Context, problemID, relatedID int64, relationType string) (*domain.RelatedLink, error) {
	relationType = strings.ToUpper(strings.TrimSpace(relationType))
	if relationType == "" {
		relationType = domain.RelationRelated
	}
	if !domain.ValidRelationType(relationType) {
		return nil, httpx.NewUnprocessable("relation_type must be one of [RELATED SIMILAR CAUSED_BY DUPLICATE WORKAROUND]")
	}
	if problemID == relatedID {
		return nil, httpx.NewUnprocessable("a problem cannot relate to itself")
	}
	if _, err := s.repo.FindByID(ctx, problemID); err != nil {
		return nil, mapRepoErr(err)
	}
	if _, err := s.repo.FindByID(ctx, relatedID); err != nil {
		return nil, httpx.NewUnprocessable("related_problem_id does not reference an existing problem")
	}

	link, err := s.related.Add(ctx, problemID, relatedID, relationType)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return link, nil
}

// DeleteRelated removes a link.
func (s *Service) DeleteRelated(ctx context.Context, problemID, relatedID int64) error {
	if err := s.related.Delete(ctx, problemID, relatedID); err != nil {
		return mapRepoErr(err)
	}
	return nil
}
