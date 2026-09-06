package server

import (
	"context"

	problemapp "github.com/team/pkb/internal/problem/application"
	tagapp "github.com/team/pkb/internal/tag/application"
)

// tagResolverAdapter bridges the tag service to the interface the problem
// service expects, so the problem feature never imports the tag application
// package directly.
type tagResolverAdapter struct {
	tags *tagapp.Service
}

func (a tagResolverAdapter) Resolve(ctx context.Context, ids []int64, names []string) ([]problemapp.ResolvedTag, error) {
	resolved, err := a.tags.Resolve(ctx, ids, names)
	if err != nil {
		return nil, err
	}
	out := make([]problemapp.ResolvedTag, len(resolved))
	for i, t := range resolved {
		out[i] = problemapp.ResolvedTag{ID: t.ID, Name: t.Name, Slug: t.Slug}
	}
	return out, nil
}
