// Package application holds the auth use cases: register, login, and lookup.
package application

import (
	"context"
	"errors"
	"strings"

	"github.com/team/pkb/internal/shared/authx"
	"github.com/team/pkb/internal/shared/httpx"
	"github.com/team/pkb/internal/user/domain"
)

// Service exposes auth operations.
type Service struct {
	repo   domain.Repository
	tokens *authx.TokenService
}

// NewService wires the auth service.
func NewService(repo domain.Repository, tokens *authx.TokenService) *Service {
	return &Service{repo: repo, tokens: tokens}
}

// RegisterInput is the payload for creating an account.
type RegisterInput struct {
	Username    string
	Email       string
	Password    string
	DisplayName string
}

// AuthResult bundles a signed token with the authenticated user.
type AuthResult struct {
	Token string
	User  domain.User
}

// Register creates a new account. The very first account in the system becomes
// an ADMIN; everyone after is a MEMBER.
func (s *Service) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	username := strings.ToLower(strings.TrimSpace(in.Username))
	email := strings.ToLower(strings.TrimSpace(in.Email))
	display := strings.TrimSpace(in.DisplayName)
	if display == "" {
		display = username
	}
	if username == "" || email == "" || len(in.Password) < 8 {
		return nil, httpx.NewUnprocessable("username, email and a password of at least 8 characters are required")
	}

	if existing, _ := s.repo.FindByLogin(ctx, username); existing != nil {
		return nil, httpx.NewConflict("username or email already registered")
	}
	if existing, _ := s.repo.FindByLogin(ctx, email); existing != nil {
		return nil, httpx.NewConflict("username or email already registered")
	}

	count, err := s.repo.CountAll(ctx)
	if err != nil {
		return nil, httpx.NewInternal(err)
	}
	role := domain.RoleMember
	if count == 0 {
		role = domain.RoleAdmin
	}

	hash, err := authx.HashPassword(in.Password)
	if err != nil {
		return nil, httpx.NewInternal(err)
	}

	u := &domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		DisplayName:  display,
		Role:         role,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		if errors.Is(err, domain.ErrCredentialsTaken) {
			return nil, httpx.NewConflict("username or email already registered")
		}
		return nil, httpx.NewInternal(err)
	}
	return s.issue(*u)
}

// Login authenticates a username-or-email + password pair.
func (s *Service) Login(ctx context.Context, login, password string) (*AuthResult, error) {
	login = strings.ToLower(strings.TrimSpace(login))
	u, err := s.repo.FindByLogin(ctx, login)
	if err != nil || u == nil {
		return nil, httpx.NewUnauthorized("invalid username or password")
	}
	if !authx.CheckPassword(u.PasswordHash, password) {
		return nil, httpx.NewUnauthorized("invalid username or password")
	}
	return s.issue(*u)
}

// Me returns the user for an id (used by GET /auth/me).
func (s *Service) Me(ctx context.Context, id int64) (*domain.User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, httpx.NewNotFound("user not found")
		}
		return nil, httpx.NewInternal(err)
	}
	return u, nil
}

func (s *Service) issue(u domain.User) (*AuthResult, error) {
	token, _, err := s.tokens.Generate(authx.Identity{
		UserID:      u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		Role:        u.Role,
	})
	if err != nil {
		return nil, httpx.NewInternal(err)
	}
	return &AuthResult{Token: token, User: u}, nil
}
