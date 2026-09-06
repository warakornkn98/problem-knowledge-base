// Package server assembles the HTTP application: it constructs every feature's
// repository -> service -> handler chain (composition root) and mounts routes.
package server

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"gorm.io/gorm"

	"github.com/team/pkb/internal/config"
	"github.com/team/pkb/internal/shared/authx"
	"github.com/team/pkb/internal/shared/httpx"

	categoryapp "github.com/team/pkb/internal/category/application"
	categoryhttp "github.com/team/pkb/internal/category/handler"
	categoryinfra "github.com/team/pkb/internal/category/infrastructure"

	tagapp "github.com/team/pkb/internal/tag/application"
	taghttp "github.com/team/pkb/internal/tag/handler"
	taginfra "github.com/team/pkb/internal/tag/infrastructure"

	userapp "github.com/team/pkb/internal/user/application"
	userhttp "github.com/team/pkb/internal/user/handler"
	userinfra "github.com/team/pkb/internal/user/infrastructure"

	problemapp "github.com/team/pkb/internal/problem/application"
	problemhttp "github.com/team/pkb/internal/problem/handler"
	probleminfra "github.com/team/pkb/internal/problem/infrastructure"

	searchapp "github.com/team/pkb/internal/search/application"
	searchhttp "github.com/team/pkb/internal/search/handler"

	dashboardapp "github.com/team/pkb/internal/dashboard/application"
	dashboardhttp "github.com/team/pkb/internal/dashboard/handler"
	dashboardinfra "github.com/team/pkb/internal/dashboard/infrastructure"
)

// New builds the fully wired Fiber application.
func New(cfg *config.Config, db *gorm.DB, log *slog.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               "Problem Knowledge Base API",
		DisableStartupMessage: cfg.IsProduction(),
		ErrorHandler:          errorHandler(log),
		BodyLimit:             4 * 1024 * 1024,
	})

	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(requestLogger(log))
	app.Use(cors.New(corsConfig(cfg.CORSOrigins)))

	// ---- Composition root: repos -> services -> handlers ----
	tokens := authx.NewTokenService(cfg.JWTSecret, cfg.JWTExpiry)
	authMW := authx.Middleware(tokens)

	categoryRepo := categoryinfra.NewRepository(db)
	categorySvc := categoryapp.NewService(categoryRepo)

	tagRepo := taginfra.NewRepository(db)
	tagSvc := tagapp.NewService(tagRepo)

	userRepo := userinfra.NewRepository(db)
	userSvc := userapp.NewService(userRepo, tokens)

	problemRepo := probleminfra.NewProblemRepository(db)
	stepRepo := probleminfra.NewStepRepository(db)
	relatedRepo := probleminfra.NewRelatedRepository(db)
	problemSvc := problemapp.NewService(
		problemRepo, stepRepo, relatedRepo,
		categoryRepo,
		tagResolverAdapter{tags: tagSvc},
	)

	searchSvc := searchapp.NewService(problemRepo)
	dashboardSvc := dashboardapp.NewService(dashboardinfra.NewRepository(db))

	// ---- Routes ----
	app.Get("/health", func(c *fiber.Ctx) error {
		return httpx.OK(c, fiber.Map{"status": "ok"})
	})

	api := app.Group("/api")

	userhttp.New(userSvc, authMW).Register(api)
	registerMeta(api)

	// Everything below requires authentication.
	protected := api.Group("", authMW)

	categoryhttp.New(categorySvc).Register(protected)
	taghttp.New(tagSvc).Register(protected)
	problemhttp.New(problemSvc).Register(protected)
	searchhttp.New(searchSvc).Register(protected)
	dashboardhttp.New(dashboardSvc).Register(protected)

	app.Use(func(c *fiber.Ctx) error {
		return httpx.Error(c, fiber.StatusNotFound, "route not found")
	})

	return app
}

func corsConfig(origins []string) cors.Config {
	joined := strings.Join(origins, ",")
	if joined == "" {
		joined = "*"
	}
	return cors.Config{
		AllowOrigins: joined,
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}
}

func requestLogger(log *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		log.Info("request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"request_id", c.Locals(requestid.ConfigDefault.ContextKey),
		)
		return err
	}
}

func errorHandler(log *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		var appErr *httpx.AppError
		if errors.As(err, &appErr) {
			if appErr.Code >= 500 {
				log.Error("request failed", "path", c.Path(), "error", appErr.Error())
			}
			return httpx.Error(c, appErr.Code, appErr.Message)
		}

		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			return httpx.Error(c, fiberErr.Code, fiberErr.Message)
		}

		log.Error("unhandled error", "path", c.Path(), "error", err.Error())
		return httpx.Error(c, fiber.StatusInternalServerError, "Internal server error")
	}
}
