// Package httpx holds the shared HTTP concerns: the response envelope, the
// domain-error to HTTP-status mapping, and pagination helpers.
package httpx

import "github.com/gofiber/fiber/v2"

// Envelope is the single response shape used across the whole API.
type Envelope struct {
	Success bool   `json:"success"`
	Data    any    `json:"data"`
	Message string `json:"message,omitempty"`
}

// Pagination describes a page of a list response.
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// List is the data payload for paginated list endpoints.
type List struct {
	Items      any        `json:"items"`
	Pagination Pagination `json:"pagination"`
}

// OK writes a 200 success envelope.
func OK(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(Envelope{Success: true, Data: data})
}

// Created writes a 201 success envelope.
func Created(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(Envelope{Success: true, Data: data})
}

// NoContentOK writes a 200 success envelope with a null data field (used for
// deletes so clients still get the standard shape).
func NoContentOK(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(Envelope{Success: true, Data: nil})
}

// Paginated writes a 200 list envelope, computing total_pages.
func Paginated(c *fiber.Ctx, items any, page, limit int, total int64) error {
	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	return c.Status(fiber.StatusOK).JSON(Envelope{
		Success: true,
		Data: List{
			Items: items,
			Pagination: Pagination{
				Page:       page,
				Limit:      limit,
				Total:      total,
				TotalPages: totalPages,
			},
		},
	})
}

// Error writes a failure envelope with the given status and message.
func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(Envelope{Success: false, Data: nil, Message: message})
}

// Fail maps any error to an *AppError and writes the failure envelope. The
// wrapped internal detail (if any) is left for the recover/logging middleware.
func Fail(c *fiber.Ctx, err error) error {
	appErr := AsAppError(err)
	return Error(c, appErr.Code, appErr.Message)
}
