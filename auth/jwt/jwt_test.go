package jwt_test

import (
	"context"
	"testing"
	"time"

	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/auth/jwt"
	"github.com/music-gang/music-gang-api/mock"
)

const jwtTestSecret = "jwt-test-secret"

func TestJWT_Exchange(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		s := jwt.NewJWTService()

		s.Secret = jwtTestSecret

		auth := &entity.Auth{
			ID:     1,
			UserID: 1,
			Source: entity.AuthSourceLocal,
			User: &entity.User{
				Name: "Jane Doe",
			},
		}

		if token, err := s.Exchange(context.Background(), auth); err != nil {
			t.Errorf("unexpected error: %v", err)
		} else if token == nil {
			t.Error("expected token, got nil")
		} else if token.AccessToken == "" || token.RefreshToken == "" {
			t.Error("expected token to have access and refresh tokens")
		}
	})

	t.Run("ErrCtxDone", func(t *testing.T) {

		s := jwt.NewJWTService()

		s.Secret = jwtTestSecret

		auth := &entity.Auth{
			ID:     1,
			UserID: 1,
			Source: entity.AuthSourceLocal,
			User: &entity.User{
				ID:   1,
				Name: "Jane Doe",
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if _, err := s.Exchange(ctx, auth); err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestJWT_Parse(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		s := jwt.NewJWTService()

		s.Secret = jwtTestSecret

		s.JWTBlacklistService = &mock.JWTBlacklistService{
			IsBlacklistedFn: func(ctx context.Context, token string) (bool, error) {
				return false, nil
			},
		}

		auth := &entity.Auth{
			ID:     1,
			UserID: 1,
			Source: entity.AuthSourceLocal,
			User: &entity.User{
				ID:   1,
				Name: "Jane Doe",
			},
		}

		token, err := s.Exchange(context.Background(), auth)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if claims, err := s.Parse(context.Background(), token.AccessToken); err != nil {
			t.Errorf("unexpected error: %v", err)
		} else if claims == nil {
			t.Error("expected claims, got nil")
		} else if claims.Auth == nil {
			t.Error("expected claims.Auth, got nil")
		} else if claims.Auth.ID != auth.ID {
			t.Errorf("expected auth ID %v, got %v", auth.ID, claims.Auth.ID)
		}
	})

	t.Run("ErrParseWrongToken", func(t *testing.T) {

		s := jwt.NewJWTService()

		s.Secret = jwtTestSecret

		s.JWTBlacklistService = &mock.JWTBlacklistService{
			IsBlacklistedFn: func(ctx context.Context, token string) (bool, error) {
				return false, nil
			},
		}

		if _, err := s.Parse(context.Background(), "wrong-token"); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("ErrBlackListedToken", func(t *testing.T) {

		s := jwt.NewJWTService()

		s.Secret = jwtTestSecret

		s.JWTBlacklistService = &mock.JWTBlacklistService{
			IsBlacklistedFn: func(ctx context.Context, token string) (bool, error) {
				return true, nil
			},
		}

		auth := &entity.Auth{
			ID:     1,
			UserID: 1,
			Source: entity.AuthSourceLocal,
			User: &entity.User{
				ID:   1,
				Name: "Jane Doe",
			},
		}

		token, err := s.Exchange(context.Background(), auth)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if _, err := s.Parse(context.Background(), token.AccessToken); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("ErrCtxDoneOnExchange", func(t *testing.T) {

		s := jwt.NewJWTService()

		s.Secret = jwtTestSecret

		s.JWTBlacklistService = &mock.JWTBlacklistService{
			IsBlacklistedFn: func(ctx context.Context, token string) (bool, error) {
				return false, nil
			},
		}

		auth := &entity.Auth{
			ID:     1,
			UserID: 1,
			Source: entity.AuthSourceLocal,
			User: &entity.User{
				ID:   1,
				Name: "Jane Doe",
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if _, err := s.Exchange(ctx, auth); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("ErrCtxDoneOnParse", func(t *testing.T) {

		s := jwt.NewJWTService()

		s.Secret = jwtTestSecret

		s.JWTBlacklistService = &mock.JWTBlacklistService{
			IsBlacklistedFn: func(ctx context.Context, token string) (bool, error) {
				return false, nil
			},
		}

		auth := &entity.Auth{
			ID:     1,
			UserID: 1,
			Source: entity.AuthSourceLocal,
			User: &entity.User{
				ID:   1,
				Name: "Jane Doe",
			},
		}

		token, err := s.Exchange(context.Background(), auth)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if _, err := s.Parse(ctx, token.AccessToken); err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestJWT_Invalidate(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		s := jwt.NewJWTService()

		s.Secret = jwtTestSecret

		s.JWTBlacklistService = &mock.JWTBlacklistService{
			IsBlacklistedFn: func(ctx context.Context, token string) (bool, error) {
				return true, nil
			},
			InvalidateFn: func(ctx context.Context, token string, expiration time.Duration) error {
				return nil
			},
		}

		auth := &entity.Auth{
			ID:     1,
			UserID: 1,
			Source: entity.AuthSourceLocal,
			User: &entity.User{
				ID:   1,
				Name: "Jane Doe",
			},
		}

		token, err := s.Exchange(context.Background(), auth)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if err := s.Invalidate(context.Background(), token.AccessToken, 24*time.Hour); err != nil {
			t.Errorf("unexpected error: %v", err)
		} else if ok, err := s.JWTBlacklistService.IsBlacklisted(context.Background(), token.AccessToken); err != nil {
			t.Errorf("unexpected error: %v", err)
		} else if !ok {
			t.Error("expected token to be blacklisted")
		}
	})
}

func TestJWT_Refresh(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		s := jwt.NewJWTService()

		s.Secret = jwtTestSecret

		s.JWTBlacklistService = &mock.JWTBlacklistService{
			IsBlacklistedFn: func(ctx context.Context, token string) (bool, error) {
				return false, nil
			},
			InvalidateFn: func(ctx context.Context, token string, expiration time.Duration) error {
				return nil
			},
		}

		auth := &entity.Auth{
			ID:     1,
			UserID: 1,
			Source: entity.AuthSourceLocal,
			User: &entity.User{
				ID:   1,
				Name: "Jane Doe",
			},
		}

		token, err := s.Exchange(context.Background(), auth)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if pair, err := s.Refresh(context.Background(), token.RefreshToken); err != nil {
			t.Errorf("unexpected error: %v", err)
		} else if pair.AccessToken == "" || pair.RefreshToken == "" {
			t.Error("expected tokens to be returned")
		}
	})

	t.Run("ErrCtxDone", func(t *testing.T) {

		s := jwt.NewJWTService()

		s.Secret = jwtTestSecret

		s.JWTBlacklistService = &mock.JWTBlacklistService{
			IsBlacklistedFn: func(ctx context.Context, token string) (bool, error) {
				return false, nil
			},
			InvalidateFn: func(ctx context.Context, token string, expiration time.Duration) error {
				return nil
			},
		}

		auth := &entity.Auth{
			ID:     1,
			UserID: 1,
			Source: entity.AuthSourceLocal,
			User: &entity.User{
				ID:   1,
				Name: "Jane Doe",
			},
		}

		token, err := s.Exchange(context.Background(), auth)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if _, err := s.Refresh(ctx, token.RefreshToken); err == nil {
			t.Error("expected error, got nil")
		}
	})
}
