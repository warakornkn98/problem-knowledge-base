// Package domain defines the User entity and its persistence port.
package domain

import (
	"context"
	"errors"
	"time"
)

// Role values.
const (
	RoleAdmin  = "ADMIN"
	RoleMember = "MEMBER"
)

// User is an account that can sign in and author problems.
type User struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
	DisplayName  string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Domain errors.
var (
	ErrNotFound          = errors.New("user not found")
	ErrCredentialsTaken  = errors.New("username or email already registered")
	ErrInvalidCredential = errors.New("invalid username or password")
)

// Repository is the persistence port for users.
type Repository interface {
	Create(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id int64) (*User, error)
	FindByLogin(ctx context.Context, login string) (*User, error) // username OR email
	CountAll(ctx context.Context) (int64, error)
}
