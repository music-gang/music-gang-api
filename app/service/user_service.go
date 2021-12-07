package service

import (
	"context"

	"github.com/music-gang/music-gang-api/app/entity"
)

type UserService interface {
	FindUserById(ctx context.Context, id int64) (*entity.User, error)

	FindUsers(ctx context.Context, filter UserFilter) (entity.Users, int, error)

	CreateUser(ctx context.Context, user *entity.User) error

	UpdateUser(ctx context.Context, user UserUpdate) (*entity.User, error)

	DeleteUser(ctx context.Context, id int64) error
}

type UserFilter struct {
	ID    *int64  `json:"id"`
	Email *string `json:"email"`

	Offset *int `json:"offset"`
	Limit  *int `json:"limit"`
}

type UserUpdate struct {
	Name *string `json:"name"`
}
