package domain

import (
	"context"
	"time"
)

// ListQuery captures every filter, the free-text term, sorting and pagination
// for the problem list / search endpoints.
type ListQuery struct {
	Search string // free text: matched against the FTS vector + trigram fallback

	Projects     []string
	CategoryIDs  []int64
	Environments []string
	Severities   []string
	Statuses     []string
	TagSlugs     []string

	CreatedFrom *time.Time
	CreatedTo   *time.Time
	SolvedFrom  *time.Time
	SolvedTo    *time.Time

	Sort   string // e.g. "-created_at", "severity", "title", "relevance"
	Limit  int
	Offset int
}

// ListResult is a page of list items plus the unpaged total.
type ListResult struct {
	Items []ListItem
	Total int64
}

// FacetCount is one bucket of a faceted aggregation.
type FacetCount struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Count int64  `json:"count"`
}

// Facets holds the aggregation of a query across its main dimensions, computed
// over the full (unpaged) result set. Powers the advanced-search sidebar.
type Facets struct {
	Categories   []FacetCount `json:"categories"`
	Statuses     []FacetCount `json:"statuses"`
	Severities   []FacetCount `json:"severities"`
	Environments []FacetCount `json:"environments"`
	Projects     []FacetCount `json:"projects"`
	Tags         []FacetCount `json:"tags"`
}

// WriteModel is the data the repository needs to create/update a problem. Tags
// and steps are handled through their own methods.
type WriteModel struct {
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
	SolvedAt     *time.Time
	ActorID      int64
}

// Repository is the persistence port for the problem aggregate.
type Repository interface {
	Create(ctx context.Context, m WriteModel) (int64, error)
	Update(ctx context.Context, id int64, m WriteModel) error
	Delete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*Problem, error)
	List(ctx context.Context, q ListQuery) (ListResult, error)
	UpdateStatus(ctx context.Context, id int64, status string, solvedAt *time.Time, actorID int64) error

	// SetTags replaces the tag set for a problem and refreshes tags_cached.
	SetTags(ctx context.Context, problemID int64, tagIDs []int64, tagsCached string) error

	// FindSimilar returns problems ranked against the given problem's own text
	// (title + error + tags), excluding itself.
	FindSimilar(ctx context.Context, problemID int64, limit int) ([]ListItem, error)

	// Facets aggregates the (unpaged) result of a query by its main dimensions.
	Facets(ctx context.Context, q ListQuery) (Facets, error)

	// DistinctProjects returns every project name currently in use.
	DistinctProjects(ctx context.Context) ([]string, error)
}

// StepDraft is an ordered (action, result) pair used for bulk replacement.
type StepDraft struct {
	Action string
	Result string
}

// StepRepository is the persistence port for troubleshooting steps.
type StepRepository interface {
	List(ctx context.Context, problemID int64) ([]Step, error)
	Append(ctx context.Context, problemID int64, action, result string) (*Step, error)
	Update(ctx context.Context, problemID, stepID int64, action, result string, stepNo *int) (*Step, error)
	Delete(ctx context.Context, problemID, stepID int64) error
	Reorder(ctx context.Context, problemID int64, orderedIDs []int64) error

	// ReplaceAll drops every step for the problem and re-inserts the given
	// drafts as steps 1..N. Empty actions are skipped by the caller.
	ReplaceAll(ctx context.Context, problemID int64, steps []StepDraft) error
}

// RelatedRepository is the persistence port for problem-to-problem links.
type RelatedRepository interface {
	List(ctx context.Context, problemID int64) ([]RelatedLink, error)
	Add(ctx context.Context, problemID, relatedID int64, relationType string) (*RelatedLink, error)
	Delete(ctx context.Context, problemID, relatedID int64) error
}
