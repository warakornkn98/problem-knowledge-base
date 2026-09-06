package authx

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/team/pkb/internal/shared/httpx"
)

const identityKey = "auth.identity"

// Middleware returns a Fiber handler that requires a valid bearer token and
// stores the resolved Identity in the request locals.
func Middleware(ts *TokenService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := bearerToken(c)
		if raw == "" {
			return httpx.Error(c, fiber.StatusUnauthorized, "Missing bearer token")
		}
		id, err := ts.Parse(raw)
		if err != nil {
			return httpx.Error(c, fiber.StatusUnauthorized, "Invalid or expired token")
		}
		c.Locals(identityKey, id)
		return c.Next()
	}
}

// RequireRole ensures the caller has one of the allowed roles. Use after Middleware.
func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *fiber.Ctx) error {
		id, ok := IdentityFrom(c)
		if !ok {
			return httpx.Error(c, fiber.StatusUnauthorized, "Authentication required")
		}
		if _, ok := allowed[id.Role]; !ok {
			return httpx.Error(c, fiber.StatusForbidden, "Insufficient permissions")
		}
		return c.Next()
	}
}

// IdentityFrom pulls the authenticated Identity out of the Fiber context.
func IdentityFrom(c *fiber.Ctx) (Identity, bool) {
	v := c.Locals(identityKey)
	id, ok := v.(Identity)
	return id, ok
}

// MustUserID returns the authenticated user id or 0 when unauthenticated.
func MustUserID(c *fiber.Ctx) int64 {
	if id, ok := IdentityFrom(c); ok {
		return id.UserID
	}
	return 0
}

func bearerToken(c *fiber.Ctx) string {
	h := c.Get(fiber.HeaderAuthorization)
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
