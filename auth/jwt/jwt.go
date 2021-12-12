package jwt

import (
	"context"

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
	Secret string
}

// NewJWTService creates a new JWT service.
func NewJWTService() *JWTService {
	return &JWTService{}
}

// Exchange a auth entity for a JWT token pair-
func (s *JWTService) Exchange(ctx context.Context, auth *entity.Auth) (*entity.Token, error) {

	accessTokenclaims := entity.NewAppClaims(auth, accessTokenExpiration)
	refreshTokenClaims := entity.NewAppClaims(auth, refreshTokenExpiration)

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenclaims)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)

	accessTokenString, err := accessToken.SignedString([]byte(s.Secret))
	if err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "failed to sign access token: %v", err)
	}

	refreshTokenString, err := refreshToken.SignedString([]byte(s.Secret))
	if err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "failed to sign refresh token: %v", err)
	}

	return &entity.Token{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		Expiry:       accessTokenclaims.ExpiresAt,
	}, nil
}

// Invalidate a JWT token.
func (s *JWTService) Invalidate(ctx context.Context, token string) error {
	return nil
}

// Parse a JWT token and return the associated claims.
func (s *JWTService) Parse(ctx context.Context, token string) (*entity.AppClaims, error) {

	t, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperr.Errorf(apperr.EINTERNAL, "unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(s.Secret), nil
	})

	if claims, ok := t.Claims.(*entity.AppClaims); ok && t.Valid {
		return claims, nil
	}

	return nil, apperr.Errorf(apperr.EINTERNAL, "failed to parse claims")
}

// Refresh a JWT token and returns the new token pair.
func (s *JWTService) Refresh(ctx context.Context, refreshToken string) (*entity.Token, error) {
	return nil, nil
}
