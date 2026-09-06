// Package handler is the HTTP transport for the dashboard.
package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/team/pkb/internal/dashboard/application"
	"github.com/team/pkb/internal/shared/httpx"
)

// Handler serves dashboard endpoints.
type Handler struct {
	svc *application.Service
}

// New builds the dashboard handler.
func New(svc *application.Service) *Handler { return &Handler{svc: svc} }

// Register mounts /dashboard routes.
func (h *Handler) Register(r fiber.Router) {
	g := r.Group("/dashboard")
	g.Get("/summary", h.summary)
	g.Get("/categories", h.categories)
	g.Get("/projects", h.projects)
	g.Get("/common-errors", h.commonErrors)
}

func (h *Handler) summary(c *fiber.Ctx) error {
	s, err := h.svc.Summary(c.Context())
	if err != nil {
		return httpx.Fail(c, httpx.NewInternal(err))
	}
	return httpx.OK(c, s)
}

func (h *Handler) categories(c *fiber.Ctx) error {
	items, err := h.svc.TopCategories(c.Context(), c.QueryInt("limit", 8))
	if err != nil {
		return httpx.Fail(c, httpx.NewInternal(err))
	}
	return httpx.OK(c, items)
}

func (h *Handler) projects(c *fiber.Ctx) error {
	items, err := h.svc.TopProjects(c.Context(), c.QueryInt("limit", 8))
	if err != nil {
		return httpx.Fail(c, httpx.NewInternal(err))
	}
	return httpx.OK(c, items)
}

func (h *Handler) commonErrors(c *fiber.Ctx) error {
	items, err := h.svc.CommonErrors(c.Context(), c.QueryInt("limit", 8))
	if err != nil {
		return httpx.Fail(c, httpx.NewInternal(err))
	}
	return httpx.OK(c, items)
}
