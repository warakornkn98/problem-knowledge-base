package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/team/pkb/internal/shared/httpx"
)

func (h *Handler) listRelated(c *fiber.Ctx) error {
	pid, err := c.ParamsInt("id")
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid id")
	}
	links, err := h.svc.ListRelated(c.Context(), int64(pid))
	if err != nil {
		return httpx.Fail(c, err)
	}
	out := make([]relatedResponse, len(links))
	for i, l := range links {
		out[i] = toRelatedResponse(l)
	}
	return httpx.OK(c, out)
}

type addRelatedRequest struct {
	RelatedProblemID int64  `json:"related_problem_id"`
	RelationType     string `json:"relation_type"`
}

func (h *Handler) addRelated(c *fiber.Ctx) error {
	pid, err := c.ParamsInt("id")
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid id")
	}
	var req addRelatedRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	link, err := h.svc.AddRelated(c.Context(), int64(pid), req.RelatedProblemID, req.RelationType)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, toRelatedResponse(*link))
}

func (h *Handler) deleteRelated(c *fiber.Ctx) error {
	pid, err := c.ParamsInt("id")
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid id")
	}
	rid, err := c.ParamsInt("relatedId")
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid related id")
	}
	if err := h.svc.DeleteRelated(c.Context(), int64(pid), int64(rid)); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.NoContentOK(c)
}
