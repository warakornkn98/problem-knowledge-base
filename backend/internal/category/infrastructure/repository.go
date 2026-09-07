// Package infrastructure is the GORM-backed adapter for the category
// Repository port.
package infrastructure

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/team/pkb/internal/category/domain"
)

// categoryModel maps the problem_categories table.
type categoryModel struct {
	ID          int64 `gorm:"primaryKey"`
	Name        string
	Slug        string
	Description string
	Color       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (categoryModel) TableName() string { return "problem_categories" }

func (m categoryModel) toDomain() domain.Category {
	return domain.Category{
		ID:          m.ID,
		Name:        m.Name,
		Slug:        m.Slug,
		Description: m.Description,
		Color:       m.Color,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// Repository is the GORM implementation of domain.Repository.
type Repository struct {
	db *gorm.DB
}

// NewRepository builds the category repository.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, c *domain.Category) error {
	m := categoryModel{Name: c.Name, Slug: c.Slug, Description: c.Description, Color: c.Color}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return translate(err)
	}
	c.ID = m.ID
	c.CreatedAt = m.CreatedAt
	c.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *Repository) Update(ctx context.Context, c *domain.Category) error {
	res := r.db.WithContext(ctx).Model(&categoryModel{}).
		Where("id = ?", c.ID).
		Updates(map[string]any{
			"name":        c.Name,
			"slug":        c.Slug,
			"description": c.Description,
			"color":       c.Color,
			"updated_at":  time.Now(),
		})
	if res.Error != nil {
		return translate(res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&categoryModel{}, id)
	if res.Error != nil {
		return translate(res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*domain.Category, error) {
	var m categoryModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	c := m.toDomain()
	return &c, nil
}

func (r *Repository) FindByName(ctx context.Context, name string) (*domain.Category, error) {
	var m categoryModel
	if err := r.db.WithContext(ctx).Where("lower(name) = lower(?)", name).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	c := m.toDomain()
	return &c, nil
}

func (r *Repository) List(ctx context.Context, withCounts bool) ([]domain.Category, error) {
	if !withCounts {
		var models []categoryModel
		if err := r.db.WithContext(ctx).Order("name ASC").Find(&models).Error; err != nil {
			return nil, err
		}
		out := make([]domain.Category, len(models))
		for i, m := range models {
			out[i] = m.toDomain()
		}
		return out, nil
	}

	// Flat scan target: GORM does not reliably flatten an embedded struct when
	// scanning an aggregate query, so list every column explicitly.
	type row struct {
		ID           int64
		Name         string
		Slug         string
		Description  string
		Color        string
		CreatedAt    time.Time
		UpdatedAt    time.Time
		ProblemCount int64
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("problem_categories AS c").
		Select("c.id, c.name, c.slug, c.description, c.color, c.created_at, c.updated_at, COUNT(p.id) AS problem_count").
		Joins("LEFT JOIN problems p ON p.category_id = c.id").
		Group("c.id").
		Order("c.name ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Category, len(rows))
	for i, rw := range rows {
		out[i] = domain.Category{
			ID:           rw.ID,
			Name:         rw.Name,
			Slug:         rw.Slug,
			Description:  rw.Description,
			Color:        rw.Color,
			CreatedAt:    rw.CreatedAt,
			UpdatedAt:    rw.UpdatedAt,
			ProblemCount: rw.ProblemCount,
		}
	}
	return out, nil
}

func (r *Repository) Exists(ctx context.Context, id int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&categoryModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *Repository) CountProblems(ctx context.Context, id int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("problems").Where("category_id = ?", id).Count(&count).Error
	return count, err
}

// translate converts driver-level unique-violation errors into domain errors.
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
