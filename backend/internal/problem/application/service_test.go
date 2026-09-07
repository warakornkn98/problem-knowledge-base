package application

import (
	"context"
	"testing"
	"time"

	"github.com/team/pkb/internal/problem/domain"
	"github.com/team/pkb/internal/shared/httpx"
)

// --- fakes ------------------------------------------------------------------

type fakeRepo struct {
	problems map[int64]*domain.Problem
	seq      int64
	tagCalls []struct {
		problemID int64
		tagIDs    []int64
		cached    string
	}
}

func newFakeRepo() *fakeRepo { return &fakeRepo{problems: map[int64]*domain.Problem{}} }

func (f *fakeRepo) Create(_ context.Context, m domain.WriteModel) (int64, error) {
	f.seq++
	f.problems[f.seq] = &domain.Problem{
		ID: f.seq, Title: m.Title, Project: m.Project, Environment: m.Environment,
		CategoryID: m.CategoryID, Severity: m.Severity, Status: m.Status, SolvedAt: m.SolvedAt,
	}
	return f.seq, nil
}

func (f *fakeRepo) Update(_ context.Context, id int64, m domain.WriteModel) error {
	p, ok := f.problems[id]
	if !ok {
		return domain.ErrNotFound
	}
	p.Title, p.Status, p.SolvedAt = m.Title, m.Status, m.SolvedAt
	return nil
}

func (f *fakeRepo) Delete(_ context.Context, id int64) error {
	if _, ok := f.problems[id]; !ok {
		return domain.ErrNotFound
	}
	delete(f.problems, id)
	return nil
}

func (f *fakeRepo) FindByID(_ context.Context, id int64) (*domain.Problem, error) {
	if p, ok := f.problems[id]; ok {
		return p, nil
	}
	return nil, domain.ErrNotFound
}

func (f *fakeRepo) List(context.Context, domain.ListQuery) (domain.ListResult, error) {
	return domain.ListResult{}, nil
}

func (f *fakeRepo) UpdateStatus(_ context.Context, id int64, status string, solvedAt *time.Time, _ int64) error {
	p, ok := f.problems[id]
	if !ok {
		return domain.ErrNotFound
	}
	p.Status, p.SolvedAt = status, solvedAt
	return nil
}

func (f *fakeRepo) SetTags(_ context.Context, problemID int64, tagIDs []int64, cached string) error {
	f.tagCalls = append(f.tagCalls, struct {
		problemID int64
		tagIDs    []int64
		cached    string
	}{problemID, tagIDs, cached})
	return nil
}

func (f *fakeRepo) FindSimilar(context.Context, int64, int) ([]domain.ListItem, error) {
	return nil, nil
}
func (f *fakeRepo) Facets(context.Context, domain.ListQuery) (domain.Facets, error) {
	return domain.Facets{}, nil
}
func (f *fakeRepo) DistinctProjects(context.Context) ([]string, error) { return nil, nil }

type fakeSteps struct{ appended int }

func (f *fakeSteps) List(context.Context, int64) ([]domain.Step, error) { return nil, nil }
func (f *fakeSteps) Append(_ context.Context, problemID int64, action, result string) (*domain.Step, error) {
	f.appended++
	return &domain.Step{ID: int64(f.appended), ProblemID: problemID, StepNo: f.appended, Action: action, Result: result}, nil
}
func (f *fakeSteps) Update(context.Context, int64, int64, string, string, *int) (*domain.Step, error) {
	return &domain.Step{}, nil
}
func (f *fakeSteps) Delete(context.Context, int64, int64) error    { return nil }
func (f *fakeSteps) Reorder(context.Context, int64, []int64) error { return nil }
func (f *fakeSteps) ReplaceAll(_ context.Context, _ int64, steps []domain.StepDraft) error {
	f.appended += len(steps)
	return nil
}

type fakeRelated struct{}

func (fakeRelated) List(context.Context, int64) ([]domain.RelatedLink, error) { return nil, nil }
func (fakeRelated) Add(context.Context, int64, int64, string) (*domain.RelatedLink, error) {
	return &domain.RelatedLink{}, nil
}
func (fakeRelated) Delete(context.Context, int64, int64) error { return nil }

type fakeCategories struct{ exists bool }

func (f fakeCategories) Exists(context.Context, int64) (bool, error) { return f.exists, nil }

type fakeTags struct{}

func (fakeTags) Resolve(_ context.Context, ids []int64, names []string) ([]ResolvedTag, error) {
	out := make([]ResolvedTag, 0, len(ids)+len(names))
	for _, id := range ids {
		out = append(out, ResolvedTag{ID: id, Name: "tag", Slug: "tag"})
	}
	for i, n := range names {
		out = append(out, ResolvedTag{ID: int64(100 + i), Name: n, Slug: n})
	}
	return out, nil
}

// --- tests ----------------------------------------------------------------

func newService(catExists bool) (*Service, *fakeRepo, *fakeSteps) {
	repo := newFakeRepo()
	steps := &fakeSteps{}
	return NewService(repo, steps, fakeRelated{}, fakeCategories{exists: catExists}, fakeTags{}), repo, steps
}

func baseInput() WriteInput {
	return WriteInput{Title: "Timeout", Project: "wazuh", Environment: "prod", CategoryID: 1}
}

func TestCreate_RequiresCoreFields(t *testing.T) {
	svc, _, _ := newService(true)
	_, err := svc.Create(context.Background(), WriteInput{Title: "only title"}, nil)
	appErr := httpx.AsAppError(err)
	if appErr.Code != 422 {
		t.Fatalf("want 422, got %d (%v)", appErr.Code, err)
	}
}

func TestCreate_RejectsUnknownCategory(t *testing.T) {
	svc, _, _ := newService(false)
	_, err := svc.Create(context.Background(), baseInput(), nil)
	if httpx.AsAppError(err).Code != 422 {
		t.Fatalf("want 422 for missing category, got %v", err)
	}
}

func TestCreate_DefaultsAndSteps(t *testing.T) {
	svc, repo, steps := newService(true)
	in := baseInput()
	in.TagNames = []string{"network", "timeout"}
	p, err := svc.Create(context.Background(), in, []StepInput{{Action: "check dns"}, {Action: ""}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Severity != domain.SeverityMedium || p.Status != domain.StatusOpen {
		t.Fatalf("defaults not applied: %+v", p)
	}
	if steps.appended != 1 {
		t.Fatalf("expected 1 non-empty step appended, got %d", steps.appended)
	}
	if len(repo.tagCalls) != 1 || repo.tagCalls[0].cached != "network timeout" {
		t.Fatalf("tags_cached not wired: %+v", repo.tagCalls)
	}
}

func TestUpdateStatus_SetsAndClearsSolvedAt(t *testing.T) {
	svc, _, _ := newService(true)
	p, _ := svc.Create(context.Background(), baseInput(), nil)

	solved, err := svc.UpdateStatus(context.Background(), p.ID, "solved", 0)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if solved.SolvedAt == nil {
		t.Fatal("expected solved_at to be set")
	}

	reopened, err := svc.UpdateStatus(context.Background(), p.ID, "open", 0)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if reopened.SolvedAt != nil {
		t.Fatal("expected solved_at cleared on re-open")
	}
}
