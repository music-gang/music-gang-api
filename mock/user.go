package mock

import (
	"context"

	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.UserService = (*UserService)(nil)

type UserService struct {
	CreateUserFn func(ctx context.Context, user *entity.User) error

	DeleteUserFn func(ctx context.Context, id int64) error

	FindUserByEmailFn func(ctx context.Context, email string) (*entity.User, error)

	FindUserByIDFn func(ctx context.Context, id int64) (*entity.User, error)

	FindUsersFn func(ctx context.Context, filter service.UserFilter) (entity.Users, int, error)

	UpdateUserFn func(ctx context.Context, id int64, user service.UserUpdate) (*entity.User, error)
}

func (s *UserService) CreateUser(ctx context.Context, user *entity.User) error {
	return s.CreateUserFn(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	return s.DeleteUserFn(ctx, id)
}

func (s *UserService) FindUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	return s.FindUserByEmailFn(ctx, email)

}

func (s *UserService) FindUserByID(ctx context.Context, id int64) (*entity.User, error) {
	return s.FindUserByIDFn(ctx, id)
}

func (s *UserService) FindUsers(ctx context.Context, filter service.UserFilter) (entity.Users, int, error) {
	return s.FindUsersFn(ctx, filter)
}

func (s *UserService) UpdateUser(ctx context.Context, id int64, user service.UserUpdate) (*entity.User, error) {
	return s.UpdateUserFn(ctx, id, user)
}
