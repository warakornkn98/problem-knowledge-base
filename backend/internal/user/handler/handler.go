// Package handler is the HTTP transport for authentication.
package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/team/pkb/internal/shared/authx"
	"github.com/team/pkb/internal/shared/httpx"
	"github.com/team/pkb/internal/shared/validatorx"
	"github.com/team/pkb/internal/user/application"
	"github.com/team/pkb/internal/user/domain"
)

// Handler serves auth endpoints.
type Handler struct {
	svc  *application.Service
	auth fiber.Handler
}

// New builds the auth handler. auth is the bearer-token middleware used to guard
// /auth/me.
func New(svc *application.Service, auth fiber.Handler) *Handler {
	return &Handler{svc: svc, auth: auth}
}

// Register mounts auth routes under /auth.
func (h *Handler) Register(r fiber.Router) {
	g := r.Group("/auth")
	g.Post("/register", h.register)
	g.Post("/login", h.login)
	g.Get("/me", h.auth, h.me)
}

type userResponse struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

func toUserResponse(u domain.User) userResponse {
	return userResponse{
		ID: u.ID, Username: u.Username, Email: u.Email,
		DisplayName: u.DisplayName, Role: u.Role,
	}
}

type registerRequest struct {
	Username    string `json:"username" validate:"required,min=3,max=50"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8,max=100"`
	DisplayName string `json:"display_name" validate:"max=100"`
}

func (h *Handler) register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	if err := validatorx.Struct(req); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.Register(c.Context(), application.RegisterInput{
		Username: req.Username, Email: req.Email,
		Password: req.Password, DisplayName: req.DisplayName,
	})
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, fiber.Map{
		"token": res.Token,
		"user":  toUserResponse(res.User),
	})
}

type loginRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (h *Handler) login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	if err := validatorx.Struct(req); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.Login(c.Context(), req.Login, req.Password)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{
		"token": res.Token,
		"user":  toUserResponse(res.User),
	})
}

func (h *Handler) me(c *fiber.Ctx) error {
	id := authx.MustUserID(c)
	u, err := h.svc.Me(c.Context(), id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, toUserResponse(*u))
}
