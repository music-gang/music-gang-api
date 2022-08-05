package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/handler"
)

// AuthLoginHandler handles the login request.
func (s *ServerAPI) AuthLoginHandler(c echo.Context) error {

	params := handler.LoginParams{}
	if err := c.Bind(&params); err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid request"), nil)
	}

	user, pair, err := s.ServiceHandler.AuthLogin(c.Request().Context(), params)
	if err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	return SuccessResponseJSON(c, http.StatusOK, tokenPairToEchoMap(pair, user))
}

// AuthLogoutHandler handles the logout request.
func (s *ServerAPI) AuthLogoutHandler(c echo.Context) error {

	pair := &entity.TokenPair{}

	if err := c.Bind(pair); err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid request"), nil)
	}

	if err := s.ServiceHandler.AuthLogout(c.Request().Context(), pair); err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	return SuccessResponseJSON(c, http.StatusOK, nil)
}

// AuthRefreshHandler handles the refresh request.
func (s *ServerAPI) AuthRefreshHandler(c echo.Context) error {

	pair := &entity.TokenPair{}

	if err := c.Bind(pair); err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid request"), nil)
	}

	user, pair, err := s.ServiceHandler.AuthRefresh(c.Request().Context(), pair)
	if err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	return SuccessResponseJSON(c, http.StatusOK, tokenPairToEchoMap(pair, user))
}

// AuthRegisterHandler handles the register request.
func (s *ServerAPI) AuthRegisterHandler(c echo.Context) error {

	params := handler.RegisterParams{}
	if err := c.Bind(&params); err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid request"), nil)
	}

	user, pair, err := s.ServiceHandler.AuthRegister(c.Request().Context(), params)
	if err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	return SuccessResponseJSON(c, http.StatusOK, tokenPairToEchoMap(pair, user))
}

// tokenPairToEchoMap converts a TokenPair to a map for JSON serialization.
func tokenPairToEchoMap(pair *entity.TokenPair, user *entity.User) echo.Map {
	return echo.Map{
		"user":          user,
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_in":    pair.Expiry,
		"token_type":    pair.TokenType,
	}
}
