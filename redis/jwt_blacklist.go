package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.JWTBlacklistService = (*JWTBlacklistService)(nil)

// JWTBlacklistService is a service for managing black list JWT tokens.
type JWTBlacklistService struct {
	db *DB
}

// NewJWTBlacklistService creates a new JWTBlacklistService.
func NewJWTBlacklistService(db *DB) *JWTBlacklistService {
	return &JWTBlacklistService{db: db}
}

// Invalidate a JWT token.
// Returns EUNAUTHORIZED if the user is not allowed to invalidate the token.
func (s *JWTBlacklistService) Invalidate(ctx context.Context, token string, expiration time.Duration) error {
	return invalidate(ctx, s.db, token, expiration)
}

// IsBlacklisted checks if a token is blacklisted.
func (s *JWTBlacklistService) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	return isBlacklisted(ctx, s.db, token)
}

// invalidate a JWT token.
func invalidate(ctx context.Context, db *DB, token string, expiration time.Duration) error {

	if err := db.client.Set(ctx, token, true, expiration).Err(); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to invalidate token: %v", err)
	}

	return nil
}

// IsBlacklisted checks if a token is blacklisted.
func isBlacklisted(ctx context.Context, db *DB, token string) (bool, error) {

	_, err := db.client.Get(ctx, token).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, apperr.Errorf(apperr.EINTERNAL, "failed to check if token is blacklisted: %v", err)
	}

	return true, nil
}
