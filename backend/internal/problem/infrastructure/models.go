// Package infrastructure is the GORM adapter for the problem aggregate ports.
package infrastructure

import (
	"time"

	"github.com/team/pkb/internal/problem/domain"
)

// problemModel maps the problems table.
type problemModel struct {
	ID           int64 `gorm:"primaryKey"`
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
	TagsCached   string
	CreatedBy    *int64
	UpdatedBy    *int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	SolvedAt     *time.Time
}

func (problemModel) TableName() string { return "problems" }

type stepModel struct {
	ID        int64 `gorm:"primaryKey"`
	ProblemID int64
	StepNo    int
	Action    string
	Result    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (stepModel) TableName() string { return "problem_steps" }

func (m stepModel) toDomain() domain.Step {
	return domain.Step{
		ID:        m.ID,
		ProblemID: m.ProblemID,
		StepNo:    m.StepNo,
		Action:    m.Action,
		Result:    m.Result,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

type relatedModel struct {
	ID               int64 `gorm:"primaryKey"`
	ProblemID        int64
	RelatedProblemID int64
	RelationType     string
	CreatedAt        time.Time
}

func (relatedModel) TableName() string { return "problem_related" }

// joinedListRow is the flat row produced by the list/search query. It carries
// the problem columns plus the joined category fields and search metadata.
type joinedListRow struct {
	ID            int64
	Title         string
	Project       string
	Environment   string
	Severity      string
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	SolvedAt      *time.Time
	CategoryID    int64
	CategoryName  string
	CategorySlug  string
	CategoryColor string
	Rank          float64
	Headline      string
}

func (r joinedListRow) toListItem() domain.ListItem {
	return domain.ListItem{
		ID:          r.ID,
		Title:       r.Title,
		Project:     r.Project,
		Environment: r.Environment,
		Severity:    r.Severity,
		Status:      r.Status,
		Category: &domain.CategoryRef{
			ID:    r.CategoryID,
			Name:  r.CategoryName,
			Slug:  r.CategorySlug,
			Color: r.CategoryColor,
		},
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
		SolvedAt:  r.SolvedAt,
		Rank:      r.Rank,
		Headline:  r.Headline,
	}
}
