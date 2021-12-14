package entity

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/music-gang/music-gang-api/app/util"
)

// AppClaims is a custom claims type for JWT
// It contains the information about the user and the standard claims
type AppClaims struct {
	jwt.StandardClaims
	Auth *Auth `json:"auth"`
}

// NewAppClaims creates a new AppClaims
func NewAppClaims(auth *Auth, expiresAfterMinutes time.Duration) *AppClaims {
	return &AppClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: util.AppNowUTC().Add(time.Duration(expiresAfterMinutes * time.Minute)).Unix(),
			NotBefore: util.AppNowUTC().Unix(),
			Subject:   fmt.Sprint(auth.UserID),
			Id:        uuid.NewString(),
			IssuedAt:  util.AppNowUTC().Unix(),
			Issuer:    "music-gang",
			Audience:  "music-gang-api",
		},
		Auth: auth,
	}
}

// TokenPair is a struct that contains the tokens and the expiration time
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Expiry       int64  `json:"expires_in"`
}
