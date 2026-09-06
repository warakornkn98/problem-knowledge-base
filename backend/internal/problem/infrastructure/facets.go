package infrastructure

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/team/pkb/internal/problem/domain"
)

type facetRow struct {
	Value string
	Label string
	Count int64
}

// Facets aggregates a query across category / status / severity / environment /
// project / tag, over the full (unpaged) match set.
func (r *ProblemRepository) Facets(ctx context.Context, q domain.ListQuery) (domain.Facets, error) {
	search := strings.TrimSpace(q.Search)
	base := func() *gorm.DB {
		return r.db.WithContext(ctx).
			Table("problems AS p").
			Joins("JOIN problem_categories c ON c.id = p.category_id").
			Scopes(filterScope(q), textScope(search))
	}

	var out domain.Facets

	// Category
	{
		var rows []facetRow
		if err := base().
			Select("c.slug AS value, c.name AS label, COUNT(*) AS count").
			Group("c.slug, c.name").
			Order("count DESC, label ASC").
			Scan(&rows).Error; err != nil {
			return out, err
		}
		out.Categories = toFacetCounts(rows)
	}

	// Single-column dimensions.
	simple := []struct {
		col    string
		target *[]domain.FacetCount
	}{
		{"p.status", &out.Statuses},
		{"p.severity", &out.Severities},
		{"p.environment", &out.Environments},
		{"p.project", &out.Projects},
	}
	for _, dim := range simple {
		var rows []facetRow
		if err := base().
			Select(dim.col + " AS value, " + dim.col + " AS label, COUNT(*) AS count").
			Group(dim.col).
			Order("count DESC, value ASC").
			Scan(&rows).Error; err != nil {
			return out, err
		}
		*dim.target = toFacetCounts(rows)
	}

	// Tags
	{
		var rows []facetRow
		if err := base().
			Joins("JOIN problem_tags pt ON pt.problem_id = p.id").
			Joins("JOIN tags t ON t.id = pt.tag_id").
			Select("t.slug AS value, t.name AS label, COUNT(*) AS count").
			Group("t.slug, t.name").
			Order("count DESC, label ASC").
			Limit(30).
			Scan(&rows).Error; err != nil {
			return out, err
		}
		out.Tags = toFacetCounts(rows)
	}

	return out, nil
}

// DistinctProjects returns every project string currently stored.
func (r *ProblemRepository) DistinctProjects(ctx context.Context) ([]string, error) {
	var out []string
	err := r.db.WithContext(ctx).
		Model(&problemModel{}).
		Distinct().
		Order("project ASC").
		Pluck("project", &out).Error
	return out, err
}

func toFacetCounts(rows []facetRow) []domain.FacetCount {
	out := make([]domain.FacetCount, len(rows))
	for i, r := range rows {
		label := r.Label
		if label == "" {
			label = r.Value
		}
		out[i] = domain.FacetCount{Value: r.Value, Label: label, Count: r.Count}
	}
	return out
}
