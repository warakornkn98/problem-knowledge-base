// Package infrastructure is the GORM adapter for the user Repository port.
package infrastructure

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/team/pkb/internal/user/domain"
)

type userModel struct {
	ID           int64 `gorm:"primaryKey"`
	Username     string
	Email        string
	PasswordHash string
	DisplayName  string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (userModel) TableName() string { return "users" }

func (m userModel) toDomain() domain.User {
	return domain.User{
		ID:           m.ID,
		Username:     m.Username,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		DisplayName:  m.DisplayName,
		Role:         m.Role,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// Repository implements domain.Repository with GORM.
type Repository struct {
	db *gorm.DB
}

// NewRepository builds the user repository.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, u *domain.User) error {
	m := userModel{
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		DisplayName:  u.DisplayName,
		Role:         u.Role,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		if isUnique(err) {
			return domain.ErrCredentialsTaken
		}
		return err
	}
	u.ID = m.ID
	u.CreatedAt = m.CreatedAt
	u.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	u := m.toDomain()
	return &u, nil
}

func (r *Repository) FindByLogin(ctx context.Context, login string) (*domain.User, error) {
	var m userModel
	err := r.db.WithContext(ctx).
		Where("lower(username) = lower(?) OR lower(email) = lower(?)", login, login).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	u := m.toDomain()
	return &u, nil
}

func (r *Repository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&userModel{}).Count(&count).Error
	return count, err
}

func isUnique(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}
