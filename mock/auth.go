package mock

import (
	"context"

	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.AuthService = (*AuthService)(nil)

type AuthService struct {
	AuthentcateFn  func(ctx context.Context, opts *entity.AuthUserOptions) (*entity.Auth, error)
	FindAuthByIDFn func(ctx context.Context, id int64) (*entity.Auth, error)
	FindAuthsFn    func(ctx context.Context, filter service.AuthFilter) (entity.Auths, int, error)
	CreateAuthFn   func(ctx context.Context, auth *entity.Auth) error
	DeleteAuthFn   func(ctx context.Context, id int64) error
}

func (a *AuthService) Auhenticate(ctx context.Context, opts *entity.AuthUserOptions) (*entity.Auth, error) {
	return a.AuthentcateFn(ctx, opts)
}

func (s *AuthService) CreateAuth(ctx context.Context, auth *entity.Auth) error {
	return s.CreateAuthFn(ctx, auth)
}

func (s *AuthService) DeleteAuth(ctx context.Context, id int64) error {
	return s.DeleteAuthFn(ctx, id)
}

func (s *AuthService) FindAuthByID(ctx context.Context, id int64) (*entity.Auth, error) {
	return s.FindAuthByIDFn(ctx, id)
}

func (s *AuthService) FindAuths(ctx context.Context, filter service.AuthFilter) (entity.Auths, int, error) {
	return s.FindAuthsFn(ctx, filter)
}
