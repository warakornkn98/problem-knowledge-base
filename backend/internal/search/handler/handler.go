// Package handler is the HTTP transport for unified search.
package handler

import (
	"github.com/gofiber/fiber/v2"

	problemhandler "github.com/team/pkb/internal/problem/handler"
	"github.com/team/pkb/internal/search/application"
	"github.com/team/pkb/internal/shared/httpx"
)

// Handler serves the search endpoint.
type Handler struct {
	svc *application.Service
}

// New builds the search handler.
func New(svc *application.Service) *Handler { return &Handler{svc: svc} }

// Register mounts /search.
func (h *Handler) Register(r fiber.Router) {
	r.Get("/search", h.search)
}

func (h *Handler) search(c *fiber.Ctx) error {
	q, page := problemhandler.ParseListQuery(c)
	if q.Sort == "" {
		q.Sort = "relevance"
	}

	res, err := h.svc.Search(c.Context(), q)
	if err != nil {
		return httpx.Fail(c, err)
	}

	totalPages := 0
	if page.Limit > 0 {
		totalPages = int((res.Total + int64(page.Limit) - 1) / int64(page.Limit))
	}

	return httpx.OK(c, fiber.Map{
		"items":  problemhandler.ToListItemResponses(res.Items),
		"facets": res.Facets,
		"pagination": httpx.Pagination{
			Page:       page.Page,
			Limit:      page.Limit,
			Total:      res.Total,
			TotalPages: totalPages,
		},
		"query": c.Query("q"),
	})
}
