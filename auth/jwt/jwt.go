package jwt

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

const (
	accessTokenExpiration  = 60           // 1 hour
	refreshTokenExpiration = 60 * 24 * 15 // 15 days
)

var _ service.JWTService = (*JWTService)(nil)

// JWTService implements the JWT service.
type JWTService struct {
	Secret              string
	JWTBlacklistService service.JWTBlacklistService
}

// NewJWTService creates a new JWT service.
func NewJWTService() *JWTService {
	return &JWTService{}
}

// Exchange a auth entity for a JWT token pair.
func (s *JWTService) Exchange(ctx context.Context, auth *entity.Auth) (*entity.TokenPair, error) {
	return exchange(ctx, auth, s.Secret)
}

// Invalidate a JWT token.
func (s *JWTService) Invalidate(ctx context.Context, token string, expiration time.Duration) error {
	return s.JWTBlacklistService.Invalidate(ctx, token, expiration)
}

// Parse a JWT token and return the associated claims.
func (s *JWTService) Parse(ctx context.Context, token string) (*entity.AppClaims, error) {
	return parse(ctx, token, s.Secret, s.JWTBlacklistService)
}

// Refresh a JWT token and returns the new token pair.
func (s *JWTService) Refresh(ctx context.Context, refreshToken string) (*entity.TokenPair, error) {

	claims, err := s.Parse(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	if err := s.Invalidate(ctx, refreshToken, refreshTokenExpiration); err != nil {
		return nil, err
	}

	return s.Exchange(ctx, claims.Auth)
}

// exchange a auth entity for a JWT token pair.
func exchange(ctx context.Context, auth *entity.Auth, secret string) (*entity.TokenPair, error) {
	select {
	case <-ctx.Done():
		return nil, apperr.Errorf(apperr.EINTERNAL, "context cancelled")
	default:
		accessTokenclaims := entity.NewAppClaims(auth, accessTokenExpiration)
		refreshTokenClaims := entity.NewAppClaims(auth, refreshTokenExpiration)

		accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenclaims)
		refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)

		accessTokenString, err := accessToken.SignedString([]byte(secret))
		if err != nil {
			return nil, apperr.Errorf(apperr.EINTERNAL, "failed to sign access token: %v", err)
		}

		refreshTokenString, err := refreshToken.SignedString([]byte(secret))
		if err != nil {
			return nil, apperr.Errorf(apperr.EINTERNAL, "failed to sign refresh token: %v", err)
		}

		return &entity.TokenPair{
			AccessToken:  accessTokenString,
			RefreshToken: refreshTokenString,
			TokenType:    "Bearer",
			Expiry:       accessTokenclaims.ExpiresAt,
		}, nil
	}
}

// parse a JWT token and return the associated claims.
func parse(ctx context.Context, token string, secret string, jwtBlacklistService service.JWTBlacklistService) (*entity.AppClaims, error) {
	select {
	case <-ctx.Done():
		return nil, apperr.Errorf(apperr.EINTERNAL, "context cancelled")
	default:
		t, err := jwt.ParseWithClaims(token, &entity.AppClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, apperr.Errorf(apperr.EINTERNAL, "unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(secret), nil
		})
		if err != nil {
			return nil, apperr.Errorf(apperr.EINTERNAL, "failed to parse token: %v", err)
		}

		claims, ok := t.Claims.(*entity.AppClaims)
		if !ok {
			return nil, apperr.Errorf(apperr.EINTERNAL, "failed to parse claims")
		}

		invalidated, err := jwtBlacklistService.IsBlacklisted(ctx, token)
		if err != nil {
			return nil, err
		}

		if !t.Valid || invalidated {
			return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "invalid token")
		}

		return claims, nil
	}

}
