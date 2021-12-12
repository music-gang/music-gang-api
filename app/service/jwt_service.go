package service

import (
	"context"

	"github.com/music-gang/music-gang-api/app/entity"
)

// JWTService is an interface for JWT service.
// It is used to generate and validate JWT tokens.
type JWTService interface {
	// Exchange a auth entity for a JWT token pair-
	Exchange(ctx context.Context, auth *entity.Auth) (*entity.Token, error)

	// Invalidate a JWT token.
	Invalidate(ctx context.Context, token string) error

	// Parse a JWT token and return the associated claims.
	Parse(ctx context.Context, token string) (*entity.AppClaims, error)

	// Refresh a JWT token and returns the new token pair.
	Refresh(ctx context.Context, refreshToken string) (*entity.Token, error)
}
