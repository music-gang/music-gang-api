package mock

import (
	"context"
	"time"

	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.JWTService = (*JWTService)(nil)

type JWTService struct {
	ExchangeFn   func(ctx context.Context, auth *entity.Auth) (*entity.TokenPair, error)
	InvalidateFn func(ctx context.Context, token string, expiration time.Duration) error
	ParseFn      func(ctx context.Context, token string) (*entity.AppClaims, error)
	RefreshFn    func(ctx context.Context, refreshToken string) (*entity.TokenPair, error)
}

func (s *JWTService) Exchange(ctx context.Context, auth *entity.Auth) (*entity.TokenPair, error) {
	if s.ExchangeFn == nil {
		panic("ExchangeFn not defined")
	}
	return s.ExchangeFn(ctx, auth)
}

func (s *JWTService) Invalidate(ctx context.Context, token string, expiration time.Duration) error {
	if s.InvalidateFn == nil {
		panic("InvalidateFn not defined")
	}
	return s.InvalidateFn(ctx, token, expiration)
}

func (s *JWTService) Parse(ctx context.Context, token string) (*entity.AppClaims, error) {
	if s.ParseFn == nil {
		panic("ParseFn not defined")
	}
	return s.ParseFn(ctx, token)
}

func (s *JWTService) Refresh(ctx context.Context, refreshToken string) (*entity.TokenPair, error) {
	if s.RefreshFn == nil {
		panic("RefreshFn not defined")
	}
	return s.RefreshFn(ctx, refreshToken)
}

var _ service.JWTBlacklistService = (*JWTBlacklistService)(nil)

type JWTBlacklistService struct {
	InvalidateFn    func(ctx context.Context, token string, expiration time.Duration) error
	IsBlacklistedFn func(ctx context.Context, token string) (bool, error)
}

func (s *JWTBlacklistService) Invalidate(ctx context.Context, token string, expiration time.Duration) error {
	if s.InvalidateFn == nil {
		panic("InvalidateFn not defined")
	}
	return s.InvalidateFn(ctx, token, expiration)
}

func (s *JWTBlacklistService) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	if s.IsBlacklistedFn == nil {
		panic("IsBlacklistedFn not defined")
	}
	return s.IsBlacklistedFn(ctx, token)
}
