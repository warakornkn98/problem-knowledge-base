package server

import (
	"github.com/gofiber/fiber/v2"

	problemdomain "github.com/team/pkb/internal/problem/domain"
	"github.com/team/pkb/internal/shared/httpx"
)

// registerMeta mounts /api/meta/enums so the frontend can build its selects
// from a single source of truth.
func registerMeta(r fiber.Router) {
	r.Get("/meta/enums", func(c *fiber.Ctx) error {
		return httpx.OK(c, fiber.Map{
			"severities":     problemdomain.Severities,
			"statuses":       problemdomain.Statuses,
			"environments":   problemdomain.Environments,
			"relation_types": problemdomain.RelationTypes,
		})
	})
}
