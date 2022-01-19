package app

import (
	"context"

	"github.com/music-gang/music-gang-api/app/entity"
)

// contextKey represents an internal key for adding context fields.
// This is considered best practice as it prevents other packages from
// interfering with our context keys.
type contextKey int

const (
	// stores the user logged in
	userContextKey = contextKey(iota + 1)
	tagContextKey

	ContextTagGeneric = "generic"
	ContextTagHTTP    = "HTTP"
	ContextTagCLI     = "CLI"
)

// NewContextWithUser returns a new context with the provided user attached.
func NewContextWithUser(ctx context.Context, user *entity.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// NewContextWithTag returns a new context with the provided tag attached.
// This can be useful during logging to define in which context a log entry was created, for example, HTTP, cron, CLI, etc.
func NewContextWithTags(ctx context.Context, tags []string) context.Context {
	return context.WithValue(ctx, tagContextKey, tags)
}

// TagsFromContext returns the tags stored in the provided context.
// If no tags are stored in the context, a slice with a single generic tag is returned.
func TagsFromContext(ctx context.Context) []string {
	if ctx == nil {
		return []string{ContextTagGeneric}
	}
	tags, ok := ctx.Value(tagContextKey).([]string)
	if !ok {
		return []string{ContextTagGeneric}
	}
	return tags
}

// UserFromContext returns the user stored in the provided context.
func UserFromContext(ctx context.Context) *entity.User {
	if ctx == nil {
		return nil
	}
	user, ok := ctx.Value(userContextKey).(*entity.User)
	if !ok {
		return nil
	}
	return user
}

// UserIDFromContext returns the user ID stored in the provided context.
func UserIDFromContext(ctx context.Context) int64 {
	if user := UserFromContext(ctx); user != nil {
		return user.ID
	}
	return 0
}
