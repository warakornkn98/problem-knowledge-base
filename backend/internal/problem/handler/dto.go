package handler

import (
	"time"

	"github.com/team/pkb/internal/problem/domain"
)

func fmtTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func fmtTimePtr(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}

type stepResponse struct {
	ID        int64  `json:"id"`
	StepNo    int    `json:"step_no"`
	Action    string `json:"action"`
	Result    string `json:"result"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toStepResponse(s domain.Step) stepResponse {
	return stepResponse{
		ID: s.ID, StepNo: s.StepNo, Action: s.Action, Result: s.Result,
		CreatedAt: fmtTime(s.CreatedAt), UpdatedAt: fmtTime(s.UpdatedAt),
	}
}

type listItemResponse struct {
	ID          int64               `json:"id"`
	Title       string              `json:"title"`
	Project     string              `json:"project"`
	Environment string              `json:"environment"`
	Severity    string              `json:"severity"`
	Status      string              `json:"status"`
	Category    *domain.CategoryRef `json:"category"`
	Tags        []domain.TagRef     `json:"tags"`
	CreatedAt   string              `json:"created_at"`
	UpdatedAt   string              `json:"updated_at"`
	SolvedAt    *string             `json:"solved_at"`
	Rank        float64             `json:"rank,omitempty"`
	Headline    string              `json:"headline,omitempty"`
}

func toListItemResponse(it domain.ListItem) listItemResponse {
	tags := it.Tags
	if tags == nil {
		tags = []domain.TagRef{}
	}
	return listItemResponse{
		ID: it.ID, Title: it.Title, Project: it.Project,
		Environment: it.Environment, Severity: it.Severity, Status: it.Status,
		Category: it.Category, Tags: tags,
		CreatedAt: fmtTime(it.CreatedAt), UpdatedAt: fmtTime(it.UpdatedAt),
		SolvedAt: fmtTimePtr(it.SolvedAt),
		Rank:     it.Rank, Headline: it.Headline,
	}
}

func toListItemResponses(items []domain.ListItem) []listItemResponse {
	out := make([]listItemResponse, len(items))
	for i, it := range items {
		out[i] = toListItemResponse(it)
	}
	return out
}

// ToListItemResponses is the exported mapper used by the search feature so both
// endpoints emit an identical problem-list shape.
func ToListItemResponses(items []domain.ListItem) any {
	return toListItemResponses(items)
}

type problemResponse struct {
	ID           int64               `json:"id"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	ErrorMessage string              `json:"error_message"`
	Category     *domain.CategoryRef `json:"category"`
	CategoryID   int64               `json:"category_id"`
	Severity     string              `json:"severity"`
	Status       string              `json:"status"`
	Environment  string              `json:"environment"`
	Project      string              `json:"project"`
	RootCause    string              `json:"root_cause"`
	Solution     string              `json:"solution"`
	Prevention   string              `json:"prevention"`
	Tags         []domain.TagRef     `json:"tags"`
	Steps        []stepResponse      `json:"steps"`
	Author       *domain.UserRef     `json:"created_by"`
	Editor       *domain.UserRef     `json:"updated_by"`
	CreatedAt    string              `json:"created_at"`
	UpdatedAt    string              `json:"updated_at"`
	SolvedAt     *string             `json:"solved_at"`
}

func toProblemResponse(p domain.Problem) problemResponse {
	tags := p.Tags
	if tags == nil {
		tags = []domain.TagRef{}
	}
	steps := make([]stepResponse, len(p.Steps))
	for i, s := range p.Steps {
		steps[i] = toStepResponse(s)
	}
	return problemResponse{
		ID: p.ID, Title: p.Title, Description: p.Description, ErrorMessage: p.ErrorMessage,
		Category: p.Category, CategoryID: p.CategoryID,
		Severity: p.Severity, Status: p.Status, Environment: p.Environment, Project: p.Project,
		RootCause: p.RootCause, Solution: p.Solution, Prevention: p.Prevention,
		Tags: tags, Steps: steps,
		Author: p.Author, Editor: p.Editor,
		CreatedAt: fmtTime(p.CreatedAt), UpdatedAt: fmtTime(p.UpdatedAt),
		SolvedAt: fmtTimePtr(p.SolvedAt),
	}
}

type relatedResponse struct {
	ID           int64             `json:"id"`
	RelationType string            `json:"relation_type"`
	CreatedAt    string            `json:"created_at"`
	Problem      *listItemResponse `json:"problem"`
}

func toRelatedResponse(l domain.RelatedLink) relatedResponse {
	r := relatedResponse{
		ID: l.ID, RelationType: l.RelationType, CreatedAt: fmtTime(l.CreatedAt),
	}
	if l.Related != nil {
		item := toListItemResponse(*l.Related)
		r.Problem = &item
	}
	return r
}
