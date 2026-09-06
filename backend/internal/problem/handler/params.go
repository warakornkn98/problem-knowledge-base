package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// queryList collects a repeatable query parameter. It accepts both
// ?tag=a&tag=b and ?tag=a,b forms and trims/deduplicates the result.
func queryList(c *fiber.Ctx, key string) []string {
	seen := map[string]bool{}
	var out []string

	add := func(raw string) {
		for _, part := range strings.Split(raw, ",") {
			part = strings.TrimSpace(part)
			if part == "" || seen[part] {
				continue
			}
			seen[part] = true
			out = append(out, part)
		}
	}

	c.Context().QueryArgs().VisitAll(func(k, v []byte) {
		if string(k) == key {
			add(string(v))
		}
	})
	return out
}

func queryInt64List(c *fiber.Ctx, key string) []int64 {
	var out []int64
	for _, s := range queryList(c, key) {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			out = append(out, n)
		}
	}
	return out
}

func queryUpperList(c *fiber.Ctx, key string) []string {
	raw := queryList(c, key)
	out := make([]string, len(raw))
	for i, s := range raw {
		out[i] = strings.ToUpper(s)
	}
	return out
}

// parseDate accepts YYYY-MM-DD or RFC3339. endOfDay pushes a bare date to the
// last instant of that day so inclusive "to" filters behave intuitively.
func parseDate(raw string, endOfDay bool) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return &t
	}
	if t, err := time.Parse("2006-01-02", raw); err == nil {
		if endOfDay {
			t = t.Add(24*time.Hour - time.Nanosecond)
		}
		return &t
	}
	return nil
}
