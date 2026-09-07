// Package application holds the problem use cases: CRUD, status changes, tag
// wiring and the similar-problem lookup. It depends only on domain ports.
package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/team/pkb/internal/problem/domain"
	"github.com/team/pkb/internal/shared/httpx"
)

// Service implements the problem use cases.
type Service struct {
	repo       domain.Repository
	steps      domain.StepRepository
	related    domain.RelatedRepository
	categories CategoryChecker
	tags       TagResolver
	now        func() time.Time
}

// NewService wires the problem service.
func NewService(
	repo domain.Repository,
	steps domain.StepRepository,
	related domain.RelatedRepository,
	categories CategoryChecker,
	tags TagResolver,
) *Service {
	return &Service{
		repo:       repo,
		steps:      steps,
		related:    related,
		categories: categories,
		tags:       tags,
		now:        time.Now,
	}
}

// StepInput is an inline step supplied on problem create.
type StepInput struct {
	Action string
	Result string
}

// WriteInput is the shared payload for Create and Update.
type WriteInput struct {
	Title        string
	Description  string
	ErrorMessage string
	CategoryID   int64
	Severity     string
	Status       string
	Environment  string
	Project      string
	RootCause    string
	Solution     string
	Prevention   string

	// Tags: nil means "leave unchanged" on update; non-nil (even empty) replaces.
	TagIDs   []int64
	TagNames []string

	ActorID int64
}

// Create validates and persists a new problem (with optional inline tags/steps).
func (s *Service) Create(ctx context.Context, in WriteInput, steps []StepInput) (*domain.Problem, error) {
	in.normalize()
	if err := s.validateWrite(ctx, in, true); err != nil {
		return nil, err
	}

	model := domain.WriteModel{
		Title:        in.Title,
		Description:  in.Description,
		ErrorMessage: in.ErrorMessage,
		CategoryID:   in.CategoryID,
		Severity:     in.Severity,
		Status:       in.Status,
		Environment:  in.Environment,
		Project:      in.Project,
		RootCause:    in.RootCause,
		Solution:     in.Solution,
		Prevention:   in.Prevention,
		ActorID:      in.ActorID,
	}
	if in.Status == domain.StatusSolved {
		now := s.now()
		model.SolvedAt = &now
	}

	id, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, httpx.NewInternal(err)
	}

	if in.TagIDs != nil || in.TagNames != nil {
		if err := s.applyTags(ctx, id, in.TagIDs, in.TagNames); err != nil {
			return nil, err
		}
	}
	for _, st := range steps {
		action := strings.TrimSpace(st.Action)
		if action == "" {
			continue
		}
		if _, err := s.steps.Append(ctx, id, action, strings.TrimSpace(st.Result)); err != nil {
			return nil, httpx.NewInternal(err)
		}
	}

	return s.Get(ctx, id)
}

// Update mutates an existing problem. When steps is non-nil the whole
// troubleshooting-step list is replaced with it (empty actions dropped); nil
// leaves the existing steps untouched.
func (s *Service) Update(ctx context.Context, id int64, in WriteInput, steps *[]StepInput) (*domain.Problem, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapRepoErr(err)
	}

	in.normalize()
	if err := s.validateWrite(ctx, in, false); err != nil {
		return nil, err
	}

	model := domain.WriteModel{
		Title:        in.Title,
		Description:  in.Description,
		ErrorMessage: in.ErrorMessage,
		CategoryID:   in.CategoryID,
		Severity:     in.Severity,
		Status:       in.Status,
		Environment:  in.Environment,
		Project:      in.Project,
		RootCause:    in.RootCause,
		Solution:     in.Solution,
		Prevention:   in.Prevention,
		ActorID:      in.ActorID,
	}

	// Keep solved_at consistent with the (possibly changed) status.
	switch {
	case in.Status == domain.StatusSolved && existing.SolvedAt != nil:
		model.SolvedAt = existing.SolvedAt
	case in.Status == domain.StatusSolved:
		now := s.now()
		model.SolvedAt = &now
	default:
		model.SolvedAt = nil
	}

	if err := s.repo.Update(ctx, id, model); err != nil {
		return nil, mapRepoErr(err)
	}

	if in.TagIDs != nil || in.TagNames != nil {
		if err := s.applyTags(ctx, id, in.TagIDs, in.TagNames); err != nil {
			return nil, err
		}
	}

	if steps != nil {
		drafts := make([]domain.StepDraft, 0, len(*steps))
		for _, st := range *steps {
			drafts = append(drafts, domain.StepDraft{Action: st.Action, Result: st.Result})
		}
		if err := s.steps.ReplaceAll(ctx, id, drafts); err != nil {
			return nil, httpx.NewInternal(err)
		}
	}

	return s.Get(ctx, id)
}

// UpdateStatus is the lightweight status-only transition used by the list view.
func (s *Service) UpdateStatus(ctx context.Context, id int64, status string, actorID int64) (*domain.Problem, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	if !domain.ValidStatus(status) {
		return nil, httpx.NewUnprocessable("status must be one of [OPEN INVESTIGATING SOLVED KNOWN]")
	}
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapRepoErr(err)
	}

	var solvedAt *time.Time
	if status == domain.StatusSolved {
		if current.SolvedAt != nil {
			solvedAt = current.SolvedAt
		} else {
			now := s.now()
			solvedAt = &now
		}
	}

	if err := s.repo.UpdateStatus(ctx, id, status, solvedAt, actorID); err != nil {
		return nil, mapRepoErr(err)
	}
	return s.Get(ctx, id)
}

// Get returns a fully hydrated problem.
func (s *Service) Get(ctx context.Context, id int64) (*domain.Problem, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return p, nil
}

// Projects returns every distinct project name in use (for filter dropdowns).
func (s *Service) Projects(ctx context.Context) ([]string, error) {
	names, err := s.repo.DistinctProjects(ctx)
	if err != nil {
		return nil, httpx.NewInternal(err)
	}
	return names, nil
}

// Delete removes a problem and its children (cascade).
func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return mapRepoErr(err)
	}
	return nil
}

// List runs a filtered / sorted / paginated query.
func (s *Service) List(ctx context.Context, q domain.ListQuery) (domain.ListResult, error) {
	res, err := s.repo.List(ctx, q)
	if err != nil {
		return domain.ListResult{}, httpx.NewInternal(err)
	}
	return res, nil
}

// Similar returns problems close to the given one.
func (s *Service) Similar(ctx context.Context, id int64, limit int) ([]domain.ListItem, error) {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return nil, mapRepoErr(err)
	}
	if limit <= 0 || limit > 20 {
		limit = 6
	}
	items, err := s.repo.FindSimilar(ctx, id, limit)
	if err != nil {
		return nil, httpx.NewInternal(err)
	}
	return items, nil
}

// --- helpers -------------------------------------------------------------

func (in *WriteInput) normalize() {
	in.Title = strings.TrimSpace(in.Title)
	in.Project = strings.TrimSpace(in.Project)
	in.Description = strings.TrimSpace(in.Description)
	in.ErrorMessage = strings.TrimRight(in.ErrorMessage, " \t\r\n")
	in.RootCause = strings.TrimSpace(in.RootCause)
	in.Solution = strings.TrimSpace(in.Solution)
	in.Prevention = strings.TrimSpace(in.Prevention)
	in.Severity = strings.ToUpper(strings.TrimSpace(in.Severity))
	in.Status = strings.ToUpper(strings.TrimSpace(in.Status))
	in.Environment = strings.ToUpper(strings.TrimSpace(in.Environment))
	if in.Severity == "" {
		in.Severity = domain.SeverityMedium
	}
	if in.Status == "" {
		in.Status = domain.StatusOpen
	}
}

func (s *Service) validateWrite(ctx context.Context, in WriteInput, _ bool) error {
	var missing []string
	if in.Title == "" {
		missing = append(missing, "title")
	}
	if in.Project == "" {
		missing = append(missing, "project")
	}
	if in.Environment == "" {
		missing = append(missing, "environment")
	}
	if in.CategoryID == 0 {
		missing = append(missing, "category_id")
	}
	if len(missing) > 0 {
		return httpx.NewUnprocessable("missing required fields: " + strings.Join(missing, ", "))
	}

	if !domain.ValidSeverity(in.Severity) {
		return httpx.NewUnprocessable("severity must be one of [LOW MEDIUM HIGH CRITICAL]")
	}
	if !domain.ValidStatus(in.Status) {
		return httpx.NewUnprocessable("status must be one of [OPEN INVESTIGATING SOLVED KNOWN]")
	}
	if !domain.ValidEnvironment(in.Environment) {
		return httpx.NewUnprocessable("environment must be one of [LOCAL DEV UAT PROD]")
	}

	ok, err := s.categories.Exists(ctx, in.CategoryID)
	if err != nil {
		return httpx.NewInternal(err)
	}
	if !ok {
		return httpx.NewUnprocessable("category_id does not reference an existing category")
	}
	return nil
}

func (s *Service) applyTags(ctx context.Context, problemID int64, ids []int64, names []string) error {
	resolved, err := s.tags.Resolve(ctx, ids, names)
	if err != nil {
		return httpx.NewInternal(err)
	}
	tagIDs := make([]int64, 0, len(resolved))
	labels := make([]string, 0, len(resolved))
	for _, t := range resolved {
		tagIDs = append(tagIDs, t.ID)
		labels = append(labels, t.Name)
	}
	if err := s.repo.SetTags(ctx, problemID, tagIDs, strings.Join(labels, " ")); err != nil {
		return httpx.NewInternal(err)
	}
	return nil
}

func mapRepoErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrNotFound):
		return httpx.NewNotFound("problem not found")
	case errors.Is(err, domain.ErrStepNotFound):
		return httpx.NewNotFound("step not found")
	case errors.Is(err, domain.ErrRelatedNotFound):
		return httpx.NewNotFound("related link not found")
	case errors.Is(err, domain.ErrSelfRelation):
		return httpx.NewUnprocessable("a problem cannot relate to itself")
	case errors.Is(err, domain.ErrRelationExists):
		return httpx.NewConflict("relation already exists")
	default:
		return httpx.NewInternal(err)
	}
}
