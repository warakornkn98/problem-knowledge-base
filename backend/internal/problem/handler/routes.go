package handler

import "github.com/gofiber/fiber/v2"

// Register mounts every problem-related route. write is the middleware chain
// required for mutating requests (auth).
func (h *Handler) Register(r fiber.Router, write ...fiber.Handler) {
	g := r.Group("/problems")

	// Read
	g.Get("/", h.list)
	g.Get("/search", h.search)
	g.Get("/:id", h.get)
	g.Get("/:id/similar", h.similar)
	g.Get("/:id/steps", h.listSteps)
	g.Get("/:id/related", h.listRelated)

	// Write
	g.Post("/", chain(write, h.create)...)
	g.Put("/:id", chain(write, h.update)...)
	g.Patch("/:id/status", chain(write, h.updateStatus)...)
	g.Delete("/:id", chain(write, h.remove)...)

	g.Post("/:id/steps", chain(write, h.addStep)...)
	g.Put("/:id/steps/reorder", chain(write, h.reorderSteps)...)
	g.Put("/:id/steps/:stepId", chain(write, h.updateStep)...)
	g.Delete("/:id/steps/:stepId", chain(write, h.deleteStep)...)

	g.Post("/:id/related", chain(write, h.addRelated)...)
	g.Delete("/:id/related/:relatedId", chain(write, h.deleteRelated)...)
}

func chain(mw []fiber.Handler, final fiber.Handler) []fiber.Handler {
	out := make([]fiber.Handler, 0, len(mw)+1)
	out = append(out, mw...)
	out = append(out, final)
	return out
}
