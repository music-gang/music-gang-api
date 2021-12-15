package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

// localSource is the source for local auth.
var localSource = entity.AuthSourceLocal

// LoginParams represents the parameters for a user authentication (local source).
type LoginParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthLogin handles the login request.
func (s *ServerAPI) AuthLogin(c echo.Context) error {

	email, password := c.FormValue("email"), c.FormValue("password")

	return handleAuthLogin(c, s.AuthService, s.JWTService, LoginParams{
		Email:    email,
		Password: password,
	})
}

// handleAuthLogin handles the login Business Logic.
func handleAuthLogin(c echo.Context, authService service.AuthService, jwtService service.JWTService, params LoginParams) error {

	auth, err := authService.Auhenticate(c.Request().Context(), &entity.AuthUserOptions{
		Source: &localSource,
		UserParams: &entity.UserParams{
			Email:    &params.Email,
			Password: &params.Password,
		},
	})
	if err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	pair, err := jwtService.Exchange(c.Request().Context(), auth)
	if err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	return SuccessResponseJSON(c, http.StatusOK, echo.Map{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_in":    pair.Expiry,
	})
}
