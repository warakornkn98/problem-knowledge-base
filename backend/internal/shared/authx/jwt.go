package authx

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Identity is the authenticated principal extracted from a token.
type Identity struct {
	UserID      int64
	Username    string
	DisplayName string
	Role        string
}

// Claims is the JWT payload.
type Claims struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	jwt.RegisteredClaims
}

// TokenService issues and verifies HS256 JWTs.
type TokenService struct {
	secret []byte
	expiry time.Duration
}

// NewTokenService builds a TokenService.
func NewTokenService(secret string, expiry time.Duration) *TokenService {
	return &TokenService{secret: []byte(secret), expiry: expiry}
}

// Generate signs a token for the given identity.
func (s *TokenService) Generate(id Identity) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.expiry)
	claims := Claims{
		Username:    id.Username,
		DisplayName: id.DisplayName,
		Role:        id.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", id.UserID),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

// Parse verifies a token and returns the identity it carries.
func (s *TokenService) Parse(raw string) (Identity, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return Identity{}, errors.New("invalid or expired token")
	}

	var uid int64
	if _, err := fmt.Sscanf(claims.Subject, "%d", &uid); err != nil {
		return Identity{}, errors.New("invalid token subject")
	}
	return Identity{
		UserID:      uid,
		Username:    claims.Username,
		DisplayName: claims.DisplayName,
		Role:        claims.Role,
	}, nil
}
