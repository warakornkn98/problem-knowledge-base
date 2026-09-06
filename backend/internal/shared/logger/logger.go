// Package logger provides a small wrapper around log/slog so the rest of the
// codebase depends on one construction point.
package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New builds a slog.Logger. In production it emits JSON, otherwise a readable
// text handler.
func New(env string) *slog.Logger {
	level := slog.LevelInfo
	if strings.EqualFold(env, "development") {
		level = slog.LevelDebug
	}

	var handler slog.Handler
	if strings.EqualFold(env, "production") {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}
	return slog.New(handler)
}
