// Package handler is the HTTP transport for the problem feature (problems,
// steps, related links, similar). All rules live in the application layer.
package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/team/pkb/internal/problem/application"
	"github.com/team/pkb/internal/problem/domain"
	"github.com/team/pkb/internal/shared/authx"
	"github.com/team/pkb/internal/shared/httpx"
)

// Handler serves the problem endpoints.
type Handler struct {
	svc *application.Service
}

// New builds the problem handler.
func New(svc *application.Service) *Handler { return &Handler{svc: svc} }

// ParseListQuery builds a domain.ListQuery from the request's query string and
// also returns the resolved pagination. Exported so the search feature can
// reuse the exact same parsing rules.
func ParseListQuery(c *fiber.Ctx) (domain.ListQuery, httpx.PageParams) {
	page := httpx.ParsePage(c)
	return domain.ListQuery{
		Search:       c.Query("q"),
		Projects:     queryList(c, "project"),
		CategoryIDs:  queryInt64List(c, "category_id"),
		Environments: queryUpperList(c, "environment"),
		Severities:   queryUpperList(c, "severity"),
		Statuses:     queryUpperList(c, "status"),
		TagSlugs:     queryList(c, "tag"),
		CreatedFrom:  parseDate(c.Query("created_from"), false),
		CreatedTo:    parseDate(c.Query("created_to"), true),
		SolvedFrom:   parseDate(c.Query("solved_from"), false),
		SolvedTo:     parseDate(c.Query("solved_to"), true),
		Sort:         c.Query("sort"),
		Limit:        page.Limit,
		Offset:       page.Offset(),
	}, page
}

func (h *Handler) list(c *fiber.Ctx) error {
	q, page := ParseListQuery(c)

	res, err := h.svc.List(c.Context(), q)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Paginated(c, toListItemResponses(res.Items), page.Page, page.Limit, res.Total)
}

// search is a thin alias of list that always sorts by relevance.
func (h *Handler) search(c *fiber.Ctx) error {
	if c.Query("sort") == "" {
		c.Request().URI().QueryArgs().Set("sort", "relevance")
	}
	return h.list(c)
}

func (h *Handler) projects(c *fiber.Ctx) error {
	names, err := h.svc.Projects(c.Context())
	if err != nil {
		return httpx.Fail(c, err)
	}
	if names == nil {
		names = []string{}
	}
	return httpx.OK(c, names)
}

func (h *Handler) get(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid id")
	}
	p, err := h.svc.Get(c.Context(), int64(id))
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, toProblemResponse(*p))
}

func (h *Handler) similar(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.Similar(c.Context(), int64(id), c.QueryInt("limit", 6))
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, toListItemResponses(items))
}

type writeRequest struct {
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	ErrorMessage string         `json:"error_message"`
	CategoryID   int64          `json:"category_id"`
	Severity     string         `json:"severity"`
	Status       string         `json:"status"`
	Environment  string         `json:"environment"`
	Project      string         `json:"project"`
	RootCause    string         `json:"root_cause"`
	Solution     string         `json:"solution"`
	Prevention   string         `json:"prevention"`
	TagIDs       []int64        `json:"tag_ids"`
	Tags         []string       `json:"tags"`
	Steps        *[]stepPayload `json:"steps"`
}

type stepPayload struct {
	Action string `json:"action"`
	Result string `json:"result"`
}

func (r writeRequest) stepInputs() []application.StepInput {
	if r.Steps == nil {
		return nil
	}
	out := make([]application.StepInput, 0, len(*r.Steps))
	for _, s := range *r.Steps {
		out = append(out, application.StepInput{Action: s.Action, Result: s.Result})
	}
	return out
}

func (r writeRequest) toInput(actorID int64) application.WriteInput {
	in := application.WriteInput{
		Title:        r.Title,
		Description:  r.Description,
		ErrorMessage: r.ErrorMessage,
		CategoryID:   r.CategoryID,
		Severity:     r.Severity,
		Status:       r.Status,
		Environment:  r.Environment,
		Project:      r.Project,
		RootCause:    r.RootCause,
		Solution:     r.Solution,
		Prevention:   r.Prevention,
		ActorID:      actorID,
	}
	// Distinguish "omitted" (nil -> leave tags untouched) from "cleared" ([]).
	if r.TagIDs != nil {
		in.TagIDs = r.TagIDs
	}
	if r.Tags != nil {
		in.TagNames = r.Tags
	}
	return in
}

func (h *Handler) create(c *fiber.Ctx) error {
	var req writeRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	actor := authx.MustUserID(c)

	steps := req.stepInputs()
	// On create, absent tag arrays still mean "no tags".
	in := req.toInput(actor)
	if req.TagIDs == nil {
		in.TagIDs = []int64{}
	}
	if req.Tags == nil {
		in.TagNames = []string{}
	}

	p, err := h.svc.Create(c.Context(), in, steps)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, toProblemResponse(*p))
}

func (h *Handler) update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid id")
	}
	var req writeRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid JSON body")
	}

	var steps *[]application.StepInput
	if req.Steps != nil {
		s := req.stepInputs()
		steps = &s
	}

	p, err := h.svc.Update(c.Context(), int64(id), req.toInput(authx.MustUserID(c)), steps)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, toProblemResponse(*p))
}

type statusRequest struct {
	Status string `json:"status"`
}

func (h *Handler) updateStatus(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid id")
	}
	var req statusRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	p, err := h.svc.UpdateStatus(c.Context(), int64(id), req.Status, authx.MustUserID(c))
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, toProblemResponse(*p))
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
