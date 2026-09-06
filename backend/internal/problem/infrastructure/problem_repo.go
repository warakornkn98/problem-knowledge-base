package infrastructure

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/team/pkb/internal/problem/domain"
)

// ProblemRepository implements domain.Repository with GORM/PostgreSQL.
type ProblemRepository struct {
	db *gorm.DB
}

// NewProblemRepository builds the problem repository.
func NewProblemRepository(db *gorm.DB) *ProblemRepository { return &ProblemRepository{db: db} }

func nullableActor(id int64) *int64 {
	if id == 0 {
		return nil
	}
	return &id
}

// Create inserts a problem row and returns its id.
func (r *ProblemRepository) Create(ctx context.Context, m domain.WriteModel) (int64, error) {
	row := problemModel{
		Title:        m.Title,
		Description:  m.Description,
		ErrorMessage: m.ErrorMessage,
		CategoryID:   m.CategoryID,
		Severity:     m.Severity,
		Status:       m.Status,
		Environment:  m.Environment,
		Project:      m.Project,
		RootCause:    m.RootCause,
		Solution:     m.Solution,
		Prevention:   m.Prevention,
		SolvedAt:     m.SolvedAt,
		CreatedBy:    nullableActor(m.ActorID),
		UpdatedBy:    nullableActor(m.ActorID),
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return 0, err
	}
	return row.ID, nil
}

// Update overwrites the mutable columns of a problem.
func (r *ProblemRepository) Update(ctx context.Context, id int64, m domain.WriteModel) error {
	res := r.db.WithContext(ctx).Model(&problemModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"title":         m.Title,
			"description":   m.Description,
			"error_message": m.ErrorMessage,
			"category_id":   m.CategoryID,
			"severity":      m.Severity,
			"status":        m.Status,
			"environment":   m.Environment,
			"project":       m.Project,
			"root_cause":    m.RootCause,
			"solution":      m.Solution,
			"prevention":    m.Prevention,
			"solved_at":     m.SolvedAt,
			"updated_by":    nullableActor(m.ActorID),
			"updated_at":    time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpdateStatus changes only status + solved_at.
func (r *ProblemRepository) UpdateStatus(ctx context.Context, id int64, status string, solvedAt *time.Time, actorID int64) error {
	res := r.db.WithContext(ctx).Model(&problemModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     status,
			"solved_at":  solvedAt,
			"updated_by": nullableActor(actorID),
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Delete removes a problem; child rows cascade via FK.
func (r *ProblemRepository) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&problemModel{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// FindByID returns a fully hydrated problem.
func (r *ProblemRepository) FindByID(ctx context.Context, id int64) (*domain.Problem, error) {
	var m problemModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	p := &domain.Problem{
		ID:           m.ID,
		Title:        m.Title,
		Description:  m.Description,
		ErrorMessage: m.ErrorMessage,
		CategoryID:   m.CategoryID,
		Severity:     m.Severity,
		Status:       m.Status,
		Environment:  m.Environment,
		Project:      m.Project,
		RootCause:    m.RootCause,
		Solution:     m.Solution,
		Prevention:   m.Prevention,
		CreatedBy:    m.CreatedBy,
		UpdatedBy:    m.UpdatedBy,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		SolvedAt:     m.SolvedAt,
	}

	if cat, err := r.loadCategory(ctx, m.CategoryID); err == nil {
		p.Category = cat
	}
	p.Tags = r.loadTags(ctx, id)
	p.Steps = r.loadSteps(ctx, id)
	p.Author = r.loadUser(ctx, m.CreatedBy)
	p.Editor = r.loadUser(ctx, m.UpdatedBy)

	return p, nil
}

// SetTags replaces the problem's tag set and refreshes tags_cached (which feeds
// the generated search_vector column).
func (r *ProblemRepository) SetTags(ctx context.Context, problemID int64, tagIDs []int64, tagsCached string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM problem_tags WHERE problem_id = ?`, problemID).Error; err != nil {
			return err
		}
		if len(tagIDs) > 0 {
			type pt struct {
				ProblemID int64
				TagID     int64
			}
			rows := make([]pt, len(tagIDs))
			for i, tid := range tagIDs {
				rows[i] = pt{ProblemID: problemID, TagID: tid}
			}
			if err := tx.Table("problem_tags").
				Clauses(clause.OnConflict{DoNothing: true}).
				Create(&rows).Error; err != nil {
				return err
			}
		}
		return tx.Model(&problemModel{}).
			Where("id = ?", problemID).
			Update("tags_cached", tagsCached).Error
	})
}

// --- hydration helpers --------------------------------------------------

func (r *ProblemRepository) loadCategory(ctx context.Context, id int64) (*domain.CategoryRef, error) {
	var ref domain.CategoryRef
	err := r.db.WithContext(ctx).
		Table("problem_categories").
		Select("id, name, slug, color").
		Where("id = ?", id).
		Scan(&ref).Error
	if err != nil || ref.ID == 0 {
		return nil, errors.New("category not found")
	}
	return &ref, nil
}

func (r *ProblemRepository) loadTags(ctx context.Context, problemID int64) []domain.TagRef {
	var tags []domain.TagRef
	_ = r.db.WithContext(ctx).
		Table("tags AS t").
		Select("t.id, t.name, t.slug").
		Joins("JOIN problem_tags pt ON pt.tag_id = t.id").
		Where("pt.problem_id = ?", problemID).
		Order("t.name ASC").
		Scan(&tags).Error
	return tags
}

func (r *ProblemRepository) loadSteps(ctx context.Context, problemID int64) []domain.Step {
	var models []stepModel
	_ = r.db.WithContext(ctx).
		Where("problem_id = ?", problemID).
		Order("step_no ASC").
		Find(&models).Error
	out := make([]domain.Step, len(models))
	for i, m := range models {
		out[i] = m.toDomain()
	}
	return out
}

func (r *ProblemRepository) loadUser(ctx context.Context, id *int64) *domain.UserRef {
	if id == nil {
		return nil
	}
	var ref domain.UserRef
	err := r.db.WithContext(ctx).
		Table("users").
		Select("id, username, display_name").
		Where("id = ?", *id).
		Scan(&ref).Error
	if err != nil || ref.ID == 0 {
		return nil
	}
	return &ref
}
