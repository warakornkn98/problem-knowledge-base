package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/team/pkb/internal/shared/httpx"
)

func (h *Handler) listSteps(c *fiber.Ctx) error {
	pid, err := c.ParamsInt("id")
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid id")
	}
	steps, err := h.svc.ListSteps(c.Context(), int64(pid))
	if err != nil {
		return httpx.Fail(c, err)
	}
	out := make([]stepResponse, len(steps))
	for i, s := range steps {
		out[i] = toStepResponse(s)
	}
	return httpx.OK(c, out)
}

type stepRequest struct {
	Action string `json:"action"`
	Result string `json:"result"`
	StepNo *int   `json:"step_no"`
}

func (h *Handler) addStep(c *fiber.Ctx) error {
	pid, err := c.ParamsInt("id")
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid id")
	}
	var req stepRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	s, err := h.svc.AddStep(c.Context(), int64(pid), req.Action, req.Result)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, toStepResponse(*s))
}

func (h *Handler) updateStep(c *fiber.Ctx) error {
	pid, err := c.ParamsInt("id")
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid id")
	}
	sid, err := c.ParamsInt("stepId")
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid step id")
	}
	var req stepRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	s, err := h.svc.UpdateStep(c.Context(), int64(pid), int64(sid), req.Action, req.Result, req.StepNo)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, toStepResponse(*s))
}

func (h *Handler) deleteStep(c *fiber.Ctx) error {
	pid, err := c.ParamsInt("id")
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid id")
	}
	sid, err := c.ParamsInt("stepId")
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid step id")
	}
	if err := h.svc.DeleteStep(c.Context(), int64(pid), int64(sid)); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.NoContentOK(c)
}

type reorderRequest struct {
	OrderedIDs []int64 `json:"ordered_ids"`
}

func (h *Handler) reorderSteps(c *fiber.Ctx) error {
	pid, err := c.ParamsInt("id")
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid id")
	}
	var req reorderRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	steps, err := h.svc.ReorderSteps(c.Context(), int64(pid), req.OrderedIDs)
	if err != nil {
		return httpx.Fail(c, err)
	}
	out := make([]stepResponse, len(steps))
	for i, s := range steps {
		out[i] = toStepResponse(s)
	}
	return httpx.OK(c, out)
}
