// Package domain holds the problem aggregate: the Problem entity, its child
// entities (Step, RelatedLink) and the persistence ports. No framework or
// database types leak in here.
package domain

import (
	"errors"
	"time"
)

// Domain errors.
var (
	ErrNotFound        = errors.New("problem not found")
	ErrInvalidEnum     = errors.New("invalid enum value")
	ErrSelfRelation    = errors.New("a problem cannot relate to itself")
	ErrRelationExists  = errors.New("relation already exists")
	ErrStepNotFound    = errors.New("step not found")
	ErrRelatedNotFound = errors.New("related link not found")
)

// TagRef is a lightweight tag reference embedded in problem views.
type TagRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// UserRef is a lightweight user reference (author/editor).
type UserRef struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

// CategoryRef is a lightweight category reference.
type CategoryRef struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Color string `json:"color"`
}

// Step is a single troubleshooting step.
type Step struct {
	ID        int64
	ProblemID int64
	StepNo    int
	Action    string
	Result    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// RelatedLink connects a problem to another problem.
type RelatedLink struct {
	ID           int64
	ProblemID    int64
	RelatedID    int64
	RelationType string
	CreatedAt    time.Time

	// Snapshot of the target problem for display.
	Related *ListItem
}

// Problem is the full aggregate root.
type Problem struct {
	ID           int64
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
	CreatedBy    *int64
	UpdatedBy    *int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	SolvedAt     *time.Time

	// Hydrated associations (filled by FindByID).
	Category *CategoryRef
	Author   *UserRef
	Editor   *UserRef
	Tags     []TagRef
	Steps    []Step
}

// ListItem is the trimmed projection used by list/search/similar responses.
type ListItem struct {
	ID          int64
	Title       string
	Project     string
	Environment string
	Severity    string
	Status      string
	Category    *CategoryRef
	Tags        []TagRef
	CreatedAt   time.Time
	UpdatedAt   time.Time
	SolvedAt    *time.Time

	// Rank/Headline are only populated by full text search.
	Rank     float64
	Headline string
}

// Validate checks the enum fields on a problem.
func (p *Problem) Validate() error {
	if !ValidSeverity(p.Severity) ||
		!ValidStatus(p.Status) ||
		!ValidEnvironment(p.Environment) {
		return ErrInvalidEnum
	}
	return nil
}

// ApplyStatusTimestamps keeps solved_at consistent with status transitions.
func (p *Problem) ApplyStatusTimestamps(now time.Time) {
	if p.Status == StatusSolved {
		if p.SolvedAt == nil {
			p.SolvedAt = &now
		}
	} else {
		p.SolvedAt = nil
	}
}
