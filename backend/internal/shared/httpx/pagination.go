package httpx

import "github.com/gofiber/fiber/v2"

// PageParams holds normalised pagination input.
type PageParams struct {
	Page  int
	Limit int
}

// Offset is the SQL OFFSET for the current page.
func (p PageParams) Offset() int { return (p.Page - 1) * p.Limit }

// ParsePage reads ?page and ?limit with sane defaults and caps.
func ParsePage(c *fiber.Ctx) PageParams {
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	limit := c.QueryInt("limit", 20)
	switch {
	case limit < 1:
		limit = 20
	case limit > 100:
		limit = 100
	}
	return PageParams{Page: page, Limit: limit}
}
