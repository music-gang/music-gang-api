package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *ServerAPI) handleLogin(c echo.Context) error {

	username, password := c.FormValue("username"), c.FormValue("password")

	_, _ = username, password

	var accessToken, refreshToken string
	var expiresIn int64

	return SuccessResponseJSON(c, http.StatusOK, echo.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    expiresIn,
	})
}
