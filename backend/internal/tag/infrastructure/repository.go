// Package infrastructure is the GORM adapter for the tag Repository port.
package infrastructure

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/team/pkb/internal/shared/slug"
	"github.com/team/pkb/internal/tag/domain"
)

type tagModel struct {
	ID        int64 `gorm:"primaryKey"`
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (tagModel) TableName() string { return "tags" }

func (m tagModel) toDomain() domain.Tag {
	return domain.Tag{
		ID:        m.ID,
		Name:      m.Name,
		Slug:      m.Slug,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// Repository implements domain.Repository with GORM.
type Repository struct {
	db *gorm.DB
}

// NewRepository builds the tag repository.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, t *domain.Tag) error {
	m := tagModel{Name: t.Name, Slug: t.Slug}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return translate(err)
	}
	t.ID = m.ID
	t.CreatedAt = m.CreatedAt
	t.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *Repository) Update(ctx context.Context, t *domain.Tag) error {
	res := r.db.WithContext(ctx).Model(&tagModel{}).
		Where("id = ?", t.ID).
		Updates(map[string]any{"name": t.Name, "slug": t.Slug, "updated_at": time.Now()})
	if res.Error != nil {
		return translate(res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&tagModel{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*domain.Tag, error) {
	var m tagModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	t := m.toDomain()
	return &t, nil
}

func (r *Repository) FindByName(ctx context.Context, name string) (*domain.Tag, error) {
	var m tagModel
	if err := r.db.WithContext(ctx).Where("lower(name) = lower(?)", name).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	t := m.toDomain()
	return &t, nil
}

func (r *Repository) List(ctx context.Context, query string, withCounts bool) ([]domain.Tag, error) {
	if !withCounts {
		q := r.db.WithContext(ctx).Model(&tagModel{}).Order("name ASC")
		if query != "" {
			q = q.Where("name ILIKE ?", "%"+query+"%")
		}
		var models []tagModel
		if err := q.Find(&models).Error; err != nil {
			return nil, err
		}
		out := make([]domain.Tag, len(models))
		for i, m := range models {
			out[i] = m.toDomain()
		}
		return out, nil
	}

	type row struct {
		tagModel
		UsageCount int64
	}
	q := r.db.WithContext(ctx).
		Table("tags AS t").
		Select("t.*, COUNT(pt.problem_id) AS usage_count").
		Joins("LEFT JOIN problem_tags pt ON pt.tag_id = t.id").
		Group("t.id").
		Order("usage_count DESC, t.name ASC")
	if query != "" {
		q = q.Where("t.name ILIKE ?", "%"+query+"%")
	}
	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Tag, len(rows))
	for i, rw := range rows {
		t := rw.tagModel.toDomain()
		t.UsageCount = rw.UsageCount
		out[i] = t
	}
	return out, nil
}

func (r *Repository) FindByIDs(ctx context.Context, ids []int64) ([]domain.Tag, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var models []tagModel
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Tag, len(models))
	for i, m := range models {
		out[i] = m.toDomain()
	}
	return out, nil
}

func (r *Repository) EnsureByNames(ctx context.Context, names []string) ([]domain.Tag, error) {
	if len(names) == 0 {
		return nil, nil
	}

	rows := make([]tagModel, len(names))
	for i, n := range names {
		rows[i] = tagModel{Name: n, Slug: slug.Make(n)}
	}

	// Insert missing rows, ignore conflicts on the unique name index.
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoNothing: true}).
		Create(&rows).Error; err != nil {
		return nil, err
	}

	var models []tagModel
	if err := r.db.WithContext(ctx).Where("name IN ?", names).Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Tag, len(models))
	for i, m := range models {
		out[i] = m.toDomain()
	}
	return out, nil
}

func translate(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint") {
		return domain.ErrNameTaken
	}
	return err
}
