package redis

import (
	"context"

	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.JWTBlacklistService = (*JWTBlacklistService)(nil)

type JWTBlacklistService struct{}

// Invalidate a JWT token.
// Returns EUNAUTHORIZED if the user is not allowed to invalidate the token.
func (s *JWTBlacklistService) Invalidate(ctx context.Context, token string) error {
	return nil
}

// IsBlacklisted checks if a token is blacklisted.
func (s *JWTBlacklistService) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	return false, nil
}
