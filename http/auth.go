package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/music-gang/music-gang-api/app/entity"
)

var localSource = entity.AuthSourceLocal

func (s *ServerAPI) handleLogin(c echo.Context) error {

	email, password := c.FormValue("email"), c.FormValue("password")

	auth, err := s.AuthService.Auhenticate(c.Request().Context(), &entity.AuthUserOptions{
		Source: &localSource,
		UserParams: &entity.UserParams{
			Email:    &email,
			Password: &password,
		},
	})
	if err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	pair, err := s.JWTService.Exchange(c.Request().Context(), auth)
	if err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	return SuccessResponseJSON(c, http.StatusOK, echo.Map{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_in":    pair.Expiry,
	})
}
