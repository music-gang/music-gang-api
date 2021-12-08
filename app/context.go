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
)

// NewContextWithUser returns a new context with the provided user attached.
func NewContextWithUser(ctx context.Context, user *entity.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
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
