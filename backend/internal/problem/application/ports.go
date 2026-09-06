package application

import "context"

// ResolvedTag is the shape the problem service needs back from the tag module.
type ResolvedTag struct {
	ID   int64
	Name string
	Slug string
}

// TagResolver turns a mix of tag ids and free-form names into concrete tags,
// creating any missing names. Implemented by an adapter over the tag service so
// the problem module never imports the tag application package directly.
type TagResolver interface {
	Resolve(ctx context.Context, ids []int64, names []string) ([]ResolvedTag, error)
}

// CategoryChecker reports whether a category id exists.
type CategoryChecker interface {
	Exists(ctx context.Context, id int64) (bool, error)
}
