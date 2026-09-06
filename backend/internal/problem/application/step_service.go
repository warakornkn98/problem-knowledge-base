package application

import (
	"context"
	"strings"

	"github.com/team/pkb/internal/problem/domain"
	"github.com/team/pkb/internal/shared/httpx"
)

// ListSteps returns the ordered troubleshooting steps for a problem.
func (s *Service) ListSteps(ctx context.Context, problemID int64) ([]domain.Step, error) {
	if _, err := s.repo.FindByID(ctx, problemID); err != nil {
		return nil, mapRepoErr(err)
	}
	steps, err := s.steps.List(ctx, problemID)
	if err != nil {
		return nil, httpx.NewInternal(err)
	}
	return steps, nil
}

// AddStep appends a step to the end of the list.
func (s *Service) AddStep(ctx context.Context, problemID int64, action, result string) (*domain.Step, error) {
	if _, err := s.repo.FindByID(ctx, problemID); err != nil {
		return nil, mapRepoErr(err)
	}
	action = strings.TrimSpace(action)
	if action == "" {
		return nil, httpx.NewUnprocessable("action is required")
	}
	step, err := s.steps.Append(ctx, problemID, action, strings.TrimSpace(result))
	if err != nil {
		return nil, httpx.NewInternal(err)
	}
	return step, nil
}

// UpdateStep changes a step's action/result and optionally its position.
func (s *Service) UpdateStep(ctx context.Context, problemID, stepID int64, action, result string, stepNo *int) (*domain.Step, error) {
	action = strings.TrimSpace(action)
	if action == "" {
		return nil, httpx.NewUnprocessable("action is required")
	}
	step, err := s.steps.Update(ctx, problemID, stepID, action, strings.TrimSpace(result), stepNo)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return step, nil
}

// DeleteStep removes a step and re-numbers the rest.
func (s *Service) DeleteStep(ctx context.Context, problemID, stepID int64) error {
	if err := s.steps.Delete(ctx, problemID, stepID); err != nil {
		return mapRepoErr(err)
	}
	return nil
}

// ReorderSteps applies a new ordering given the full list of step ids.
func (s *Service) ReorderSteps(ctx context.Context, problemID int64, orderedIDs []int64) ([]domain.Step, error) {
	if _, err := s.repo.FindByID(ctx, problemID); err != nil {
		return nil, mapRepoErr(err)
	}
	if len(orderedIDs) == 0 {
		return nil, httpx.NewUnprocessable("ordered_ids is required")
	}
	if err := s.steps.Reorder(ctx, problemID, orderedIDs); err != nil {
		return nil, mapRepoErr(err)
	}
	return s.ListSteps(ctx, problemID)
}
