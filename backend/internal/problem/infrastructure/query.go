package infrastructure

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/team/pkb/internal/problem/domain"
)

// selectColumns is the projection shared by list and similar queries.
const selectColumns = `p.id, p.title, p.project, p.environment, p.severity, p.status,
	p.created_at, p.updated_at, p.solved_at,
	c.id AS category_id, c.name AS category_name, c.slug AS category_slug, c.color AS category_color`

// filterScope applies every non-text filter from a ListQuery.
func filterScope(q domain.ListQuery) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(q.Projects) > 0 {
			db = db.Where("p.project IN ?", q.Projects)
		}
		if len(q.CategoryIDs) > 0 {
			db = db.Where("p.category_id IN ?", q.CategoryIDs)
		}
		if len(q.Environments) > 0 {
			db = db.Where("p.environment IN ?", q.Environments)
		}
		if len(q.Severities) > 0 {
			db = db.Where("p.severity IN ?", q.Severities)
		}
		if len(q.Statuses) > 0 {
			db = db.Where("p.status IN ?", q.Statuses)
		}
		if len(q.TagSlugs) > 0 {
			db = db.Where(`EXISTS (
				SELECT 1 FROM problem_tags pt
				JOIN tags t ON t.id = pt.tag_id
				WHERE pt.problem_id = p.id AND t.slug IN ?)`, q.TagSlugs)
		}
		if q.CreatedFrom != nil {
			db = db.Where("p.created_at >= ?", *q.CreatedFrom)
		}
		if q.CreatedTo != nil {
			db = db.Where("p.created_at <= ?", *q.CreatedTo)
		}
		if q.SolvedFrom != nil {
			db = db.Where("p.solved_at >= ?", *q.SolvedFrom)
		}
		if q.SolvedTo != nil {
			db = db.Where("p.solved_at <= ?", *q.SolvedTo)
		}
		return db
	}
}

// textScope applies the free-text predicate (FTS vector OR trigram/ILIKE fallback).
func textScope(search string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		search = strings.TrimSpace(search)
		if search == "" {
			return db
		}
		like := "%" + search + "%"
		return db.Where(`(
			p.search_vector @@ websearch_to_tsquery('english', ?)
			OR p.title ILIKE ?
			OR p.error_message ILIKE ?
			OR p.tags_cached ILIKE ?
			OR p.project ILIKE ?
		)`, search, like, like, like, like)
	}
}

func orderExpression(sort, search string) string {
	switch sort {
	case "created_at":
		return "p.created_at ASC"
	case "-created_at":
		return "p.created_at DESC"
	case "updated_at":
		return "p.updated_at ASC"
	case "-updated_at":
		return "p.updated_at DESC"
	case "title":
		return "p.title ASC"
	case "-title":
		return "p.title DESC"
	case "severity", "-severity":
		return severityOrder + ", p.created_at DESC"
	case "relevance":
		if search != "" {
			return "rank DESC, p.created_at DESC"
		}
		return "p.created_at DESC"
	}
	if search != "" {
		return "rank DESC, p.created_at DESC"
	}
	return "p.created_at DESC"
}

const severityOrder = `CASE p.severity
	WHEN 'CRITICAL' THEN 4 WHEN 'HIGH' THEN 3
	WHEN 'MEDIUM' THEN 2 WHEN 'LOW' THEN 1 ELSE 0 END DESC`

// List runs the filtered / sorted / paginated problem query.
func (r *ProblemRepository) List(ctx context.Context, q domain.ListQuery) (domain.ListResult, error) {
	search := strings.TrimSpace(q.Search)

	// ---- total (filters only, no pagination) ----
	var total int64
	countTx := r.db.WithContext(ctx).
		Table("problems AS p").
		Joins("JOIN problem_categories c ON c.id = p.category_id").
		Scopes(filterScope(q), textScope(search))
	if err := countTx.Count(&total).Error; err != nil {
		return domain.ListResult{}, err
	}
	if total == 0 {
		return domain.ListResult{Items: []domain.ListItem{}, Total: 0}, nil
	}

	// ---- page ----
	selectExpr := selectColumns
	var selectArgs []any
	if search != "" {
		// Highlight markers are neutral sentinels; the API/DTO layer HTML-escapes
		// the fragment and only then swaps them for <mark> tags.
		selectExpr += `,
			ts_rank(p.search_vector, websearch_to_tsquery('english', ?)) AS rank,
			ts_headline('english',
				coalesce(nullif(p.error_message, ''), nullif(p.description, ''), p.title),
				websearch_to_tsquery('english', ?),
				'StartSel=[[hl]],StopSel=[[/hl]],MaxFragments=2,MinWords=4,MaxWords=18'
			) AS headline`
		selectArgs = append(selectArgs, search, search)
	}

	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}

	var rows []joinedListRow
	tx := r.db.WithContext(ctx).
		Table("problems AS p").
		Joins("JOIN problem_categories c ON c.id = p.category_id").
		Scopes(filterScope(q), textScope(search)).
		Select(selectExpr, selectArgs...).
		Order(orderExpression(q.Sort, search)).
		Limit(limit).
		Offset(q.Offset)

	if err := tx.Scan(&rows).Error; err != nil {
		return domain.ListResult{}, err
	}

	items := make([]domain.ListItem, len(rows))
	ids := make([]int64, len(rows))
	for i, row := range rows {
		items[i] = row.toListItem()
		ids[i] = row.ID
	}
	r.attachTags(ctx, items, ids)

	return domain.ListResult{Items: items, Total: total}, nil
}

// FindSimilar suggests related problems using the signals the spec calls for:
// shared tags, same category, same project, and a full-text overlap on the
// source problem's title + error message. Results are scored by a weighted sum
// of those signals so a partial match still surfaces.
func (r *ProblemRepository) FindSimilar(ctx context.Context, problemID int64, limit int) ([]domain.ListItem, error) {
	var src struct {
		CategoryID   int64
		Project      string
		Title        string
		ErrorMessage string
	}
	if err := r.db.WithContext(ctx).
		Table("problems").
		Select("category_id, project, title, error_message").
		Where("id = ?", problemID).
		Scan(&src).Error; err != nil {
		return nil, err
	}
	if src.CategoryID == 0 {
		return []domain.ListItem{}, nil
	}

	args := map[string]any{
		"src":  problemID,
		"cat":  src.CategoryID,
		"proj": src.Project,
		"seed": nullSeed(strings.TrimSpace(src.Title + " " + src.ErrorMessage)),
	}

	var rows []joinedListRow
	err := r.db.WithContext(ctx).
		Table("problems AS p").
		Joins("JOIN problem_categories c ON c.id = p.category_id").
		Select(selectColumns+`,
			(
				3 * (SELECT count(*) FROM problem_tags pt
				     WHERE pt.problem_id = p.id
				       AND pt.tag_id IN (SELECT tag_id FROM problem_tags WHERE problem_id = @src))
				+ CASE WHEN p.category_id = @cat THEN 2 ELSE 0 END
				+ CASE WHEN p.project = @proj THEN 2 ELSE 0 END
				+ 4 * ts_rank(p.search_vector, websearch_to_tsquery('english', @seed))
			) AS rank,
			'' AS headline`, args).
		Where("p.id <> @src", args).
		Where(`(
			p.category_id = @cat
			OR p.project = @proj
			OR p.search_vector @@ websearch_to_tsquery('english', @seed)
			OR EXISTS (SELECT 1 FROM problem_tags pt
			           WHERE pt.problem_id = p.id
			             AND pt.tag_id IN (SELECT tag_id FROM problem_tags WHERE problem_id = @src))
		)`, args).
		Order("rank DESC, p.created_at DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]domain.ListItem, len(rows))
	ids := make([]int64, len(rows))
	for i, row := range rows {
		items[i] = row.toListItem()
		items[i].Rank = 0 // internal score, not meaningful to clients
		ids[i] = row.ID
	}
	r.attachTags(ctx, items, ids)
	return items, nil
}

// nullSeed avoids passing an empty string to websearch_to_tsquery (which would
// just never match); a single space is harmless.
func nullSeed(s string) string {
	if s == "" {
		return " "
	}
	return s
}

// attachTags batch-loads tags for a set of problems and assigns them onto items.
func (r *ProblemRepository) attachTags(ctx context.Context, items []domain.ListItem, ids []int64) {
	if len(ids) == 0 {
		return
	}
	type tagRow struct {
		ProblemID int64
		ID        int64
		Name      string
		Slug      string
	}
	var tagRows []tagRow
	_ = r.db.WithContext(ctx).
		Table("problem_tags AS pt").
		Select("pt.problem_id, t.id, t.name, t.slug").
		Joins("JOIN tags t ON t.id = pt.tag_id").
		Where("pt.problem_id IN ?", ids).
		Order("t.name ASC").
		Scan(&tagRows).Error

	byProblem := make(map[int64][]domain.TagRef, len(ids))
	for _, tr := range tagRows {
		byProblem[tr.ProblemID] = append(byProblem[tr.ProblemID], domain.TagRef{
			ID: tr.ID, Name: tr.Name, Slug: tr.Slug,
		})
	}
	for i := range items {
		items[i].Tags = byProblem[items[i].ID]
		if items[i].Tags == nil {
			items[i].Tags = []domain.TagRef{}
		}
	}
}
