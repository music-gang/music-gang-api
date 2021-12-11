package service

import (
	"context"

	"github.com/music-gang/music-gang-api/app/entity"
)

// UserService rapresents the user managment service.
type UserService interface {
	// CreateUser creates a new user.
	CreateUser(ctx context.Context, user *entity.User) error

	// DeleteUser deletes the user with the given id.
	// Return EUNAUTHORIZED if the user is not the same as the authenticated user.
	// Return ENOTFOUND if the user does not exist.
	DeleteUser(ctx context.Context, id int64) error

	// FindUserByEmail returns the user with the given email.
	// Return ENOTFOUND if the user does not exist.
	FindUserByEmail(ctx context.Context, email string) (*entity.User, error)

	// FindUserByID returns the user with the given id.
	// Return ENOTFOUND if the user does not exist.
	FindUserByID(ctx context.Context, id int64) (*entity.User, error)

	// FindUsers returns a list of users filtered by the given options.
	// Also returns the total count of auths.
	FindUsers(ctx context.Context, filter UserFilter) (entity.Users, int, error)

	// UpdateUser updates the given user.
	// Return EUNAUTHORIZED if the user is not the same as the authenticated user.
	// Return ENOTFOUND if the user does not exist.
	UpdateUser(ctx context.Context, id int64, user UserUpdate) (*entity.User, error)
}

// UserFilter represents the options used to filter the users.
type UserFilter struct {
	ID    *int64  `json:"id"`
	Email *string `json:"email"`
	Name  *string `json:"name"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type UserUpdate struct {
	Name *string `json:"name"`
}
