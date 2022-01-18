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

func (s *AuthService) Auhenticate(ctx context.Context, opts *entity.AuthUserOptions) (*entity.Auth, error) {
	if s.AuthentcateFn == nil {
		panic("AuthentcateFn is not defined")
	}
	return s.AuthentcateFn(ctx, opts)
}

func (s *AuthService) CreateAuth(ctx context.Context, auth *entity.Auth) error {
	if s.CreateAuthFn == nil {
		panic("CreateAuthFn is not defined")
	}
	return s.CreateAuthFn(ctx, auth)
}

func (s *AuthService) DeleteAuth(ctx context.Context, id int64) error {
	if s.DeleteAuthFn == nil {
		panic("DeleteAuthFn is not defined")
	}
	return s.DeleteAuthFn(ctx, id)
}

func (s *AuthService) FindAuthByID(ctx context.Context, id int64) (*entity.Auth, error) {
	if s.FindAuthByIDFn == nil {
		panic("FindAuthByIDFn is not defined")
	}
	return s.FindAuthByIDFn(ctx, id)
}

func (s *AuthService) FindAuths(ctx context.Context, filter service.AuthFilter) (entity.Auths, int, error) {
	if s.FindAuthsFn == nil {
		panic("FindAuthsFn is not defined")
	}
	return s.FindAuthsFn(ctx, filter)
}
