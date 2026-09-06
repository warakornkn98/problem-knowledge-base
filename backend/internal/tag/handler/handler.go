// Package handler is the HTTP transport for tags.
package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/team/pkb/internal/shared/httpx"
	"github.com/team/pkb/internal/tag/application"
	"github.com/team/pkb/internal/tag/domain"
)

// Handler serves tag endpoints.
type Handler struct {
	svc *application.Service
}

// New builds the tag handler.
func New(svc *application.Service) *Handler { return &Handler{svc: svc} }

// Register mounts tag routes.
func (h *Handler) Register(r fiber.Router, write ...fiber.Handler) {
	g := r.Group("/tags")
	g.Get("/", h.list)
	g.Post("/", append(write, h.create)...)
	g.Put("/:id", append(write, h.update)...)
	g.Delete("/:id", append(write, h.remove)...)
}

type tagResponse struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	UsageCount int64  `json:"usage_count"`
}

func toResponse(t domain.Tag) tagResponse {
	return tagResponse{ID: t.ID, Name: t.Name, Slug: t.Slug, UsageCount: t.UsageCount}
}

func (h *Handler) list(c *fiber.Ctx) error {
	items, err := h.svc.List(c.Context(), c.Query("q"), c.QueryBool("with_counts", true))
	if err != nil {
		return httpx.Fail(c, err)
	}
	out := make([]tagResponse, len(items))
	for i, it := range items {
		out[i] = toResponse(it)
	}
	return httpx.OK(c, out)
}

type upsertRequest struct {
	Name string `json:"name"`
}

func (h *Handler) create(c *fiber.Ctx) error {
	var req upsertRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	t, err := h.svc.Create(c.Context(), req.Name)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, toResponse(*t))
}

func (h *Handler) update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid id")
	}
	var req upsertRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	t, err := h.svc.Update(c.Context(), int64(id), req.Name)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, toResponse(*t))
}

func (h *Handler) remove(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid id")
	}
	if err := h.svc.Delete(c.Context(), int64(id)); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.NoContentOK(c)
}
