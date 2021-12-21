package http

import (
	"net/http"
	"net/mail"

	"github.com/labstack/echo/v4"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/util"
	"gopkg.in/guregu/null.v4"
)

// localSource is the source for local auth.
var localSource = entity.AuthSourceLocal

// LoginParams represents the parameters for a user authentication (local source).
type LoginParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// validate is the validation function for the LoginParams.
func (p *LoginParams) validate() error {

	if p.Email == "" {
		return apperr.Errorf(apperr.EINVALID, "email is required")
	} else if _, err := mail.ParseAddress(p.Email); err != nil {
		return apperr.Errorf(apperr.EINVALID, "email is invalid")
	}

	if p.Password == "" {
		return apperr.Errorf(apperr.EINVALID, "password is required")
	}

	return nil
}

// RegisterParams represents the parameters for a user registration (local source).
type RegisterParams struct {
	Email           string `json:"email"`
	Name            string `json:"name"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

// validate is the validation function for the RegisterParams.
func (p *RegisterParams) validate() error {

	if p.Email == "" {
		return apperr.Errorf(apperr.EINVALID, "email is required")
	} else if _, err := mail.ParseAddress(p.Email); err != nil {
		return apperr.Errorf(apperr.EINVALID, "email is invalid")
	}

	if p.Name == "" {
		return apperr.Errorf(apperr.EINVALID, "name is required")
	}

	if p.Password == "" {
		return apperr.Errorf(apperr.EINVALID, "password is required")
	} else if ok := util.IsValidPassword(p.Password); !ok {
		return apperr.Errorf(apperr.EINVALID, util.PasswordRequirements)
	}

	if p.Password != p.ConfirmPassword {
		return apperr.Errorf(apperr.EINVALID, "passwords do not match")
	}

	return nil
}

// AuthLogin handles the login request.
func (s *ServerAPI) AuthLogin(c echo.Context) error {

	params := LoginParams{}
	if err := c.Bind(&params); err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid request"), nil)
	}

	if err := params.validate(); err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	return handleAuthLogin(c, s, params)
}

// AuthLogout handles the logout request.
func (s *ServerAPI) AuthLogout(c echo.Context) error {

	pair := &entity.TokenPair{}

	if err := c.Bind(pair); err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid request"), nil)
	}

	return handleAuthLogout(c, s, pair)
}

// AuthRefresh handles the refresh request.
func (s *ServerAPI) AuthRefresh(c echo.Context) error {

	pair := &entity.TokenPair{}

	if err := c.Bind(pair); err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid request"), nil)
	}

	return handleAuthRefresh(c, s, pair)
}

// AuthRegister handles the register request.
func (s *ServerAPI) AuthRegister(c echo.Context) error {

	params := RegisterParams{}
	if err := c.Bind(&params); err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid request"), nil)
	}

	if err := params.validate(); err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	return handleAuthRegister(c, s, params)
}

// handleAuthLogin handles the login Business Logic.
func handleAuthLogin(c echo.Context, server *ServerAPI, params LoginParams) error {

	auth, err := server.AuthService.Auhenticate(c.Request().Context(), &entity.AuthUserOptions{
		Source: &localSource,
		UserParams: &entity.UserParams{
			Email:    &params.Email,
			Password: &params.Password,
		},
	})
	if err != nil {
		if apperr.ErrorCode(err) == apperr.ENOTFOUND {
			return ErrorResponseJSON(c, apperr.Errorf(apperr.EUNAUTHORIZED, "wrong credentials"), nil)
		}
		return ErrorResponseJSON(c, err, nil)
	}

	if auth.User.Auths != nil {
		auth.User.Auths = nil
	}

	pair, err := server.JWTService.Exchange(c.Request().Context(), auth)
	if err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	return SuccessResponseJSON(c, http.StatusOK, tokenPairToEchoMap(pair))
}

// handleAuthLogout handles the logout Business Logic.
func handleAuthLogout(c echo.Context, server *ServerAPI, pair *entity.TokenPair) error {

	if pair.AccessToken != "" {
		if err := server.JWTService.Invalidate(c.Request().Context(), pair.AccessToken, entity.AccessTokenExpiration); err != nil {
			return ErrorResponseJSON(c, err, nil)
		}
	}

	if pair.RefreshToken != "" {
		if err := server.JWTService.Invalidate(c.Request().Context(), pair.RefreshToken, entity.RefreshTokenExpiration); err != nil {
			return ErrorResponseJSON(c, err, nil)
		}
	}

	return SuccessResponseJSON(c, http.StatusOK, nil)
}

// handleAuthRefresh handles the refresh Business Logic.
func handleAuthRefresh(c echo.Context, server *ServerAPI, pair *entity.TokenPair) error {

	if pair.RefreshToken == "" {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "refresh token is required"), nil)
	}

	pair, err := server.JWTService.Refresh(c.Request().Context(), pair.RefreshToken)
	if err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	return SuccessResponseJSON(c, http.StatusOK, tokenPairToEchoMap(pair))
}

// handleAuthRegister handles the register Business Logic.
// On success, the user is created and the JWT pairs is returned.
func handleAuthRegister(c echo.Context, server *ServerAPI, params RegisterParams) error {

	passwordhashed, err := util.HashPassword(params.Password)
	if err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	if err := server.AuthService.CreateAuth(c.Request().Context(), &entity.Auth{
		Source: localSource,
		User: &entity.User{
			Email:    null.StringFrom(params.Email),
			Name:     params.Name,
			Password: null.StringFrom(string(passwordhashed)),
		},
	}); err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	return handleAuthLogin(c, server, LoginParams{
		Email:    params.Email,
		Password: params.Password,
	})
}

// tokenPairToEchoMap converts a TokenPair to a map for JSON serialization.
func tokenPairToEchoMap(pair *entity.TokenPair) echo.Map {
	return echo.Map{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_in":    pair.Expiry,
	}
}
