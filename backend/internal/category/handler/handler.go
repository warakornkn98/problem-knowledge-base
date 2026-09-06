// Package handler is the HTTP transport for the category feature. It only maps
// requests/responses; all rules live in the application layer.
package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/team/pkb/internal/category/application"
	"github.com/team/pkb/internal/category/domain"
	"github.com/team/pkb/internal/shared/httpx"
)

// Handler serves category endpoints.
type Handler struct {
	svc *application.Service
}

// New builds the category handler.
func New(svc *application.Service) *Handler { return &Handler{svc: svc} }

// Register mounts routes under the given router (already namespaced at /api).
// write is the middleware chain required for mutating requests.
func (h *Handler) Register(r fiber.Router, write ...fiber.Handler) {
	g := r.Group("/categories")
	g.Get("/", h.list)
	g.Get("/:id", h.get)
	g.Post("/", append(write, h.create)...)
	g.Put("/:id", append(write, h.update)...)
	g.Delete("/:id", append(write, h.remove)...)
}

type categoryResponse struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Description  string `json:"description"`
	Color        string `json:"color"`
	ProblemCount int64  `json:"problem_count"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func toResponse(c domain.Category) categoryResponse {
	return categoryResponse{
		ID:           c.ID,
		Name:         c.Name,
		Slug:         c.Slug,
		Description:  c.Description,
		Color:        c.Color,
		ProblemCount: c.ProblemCount,
		CreatedAt:    c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *Handler) list(c *fiber.Ctx) error {
	withCounts := c.QueryBool("with_counts", true)
	items, err := h.svc.List(c.Context(), withCounts)
	if err != nil {
		return respondErr(c, err)
	}
	out := make([]categoryResponse, len(items))
	for i, it := range items {
		out[i] = toResponse(it)
	}
	return httpx.OK(c, out)
}

func (h *Handler) get(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid id")
	}
	cat, err := h.svc.Get(c.Context(), int64(id))
	if err != nil {
		return respondErr(c, err)
	}
	return httpx.OK(c, toResponse(*cat))
}

type upsertRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

func (h *Handler) create(c *fiber.Ctx) error {
	var req upsertRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	cat, err := h.svc.Create(c.Context(), application.CreateInput{
		Name: req.Name, Description: req.Description, Color: req.Color,
	})
	if err != nil {
		return respondErr(c, err)
	}
	return httpx.Created(c, toResponse(*cat))
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
	cat, err := h.svc.Update(c.Context(), int64(id), application.UpdateInput{
		Name: req.Name, Description: req.Description, Color: req.Color,
	})
	if err != nil {
		return respondErr(c, err)
	}
	return httpx.OK(c, toResponse(*cat))
}

func (h *Handler) remove(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid id")
	}
	if err := h.svc.Delete(c.Context(), int64(id)); err != nil {
		return respondErr(c, err)
	}
	return httpx.NoContentOK(c)
}

func respondErr(c *fiber.Ctx, err error) error {
	appErr := httpx.AsAppError(err)
	return httpx.Error(c, appErr.Code, appErr.Message)
}
