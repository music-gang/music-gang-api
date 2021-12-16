package http

import (
	"fmt"
	"net/http"
	"net/mail"

	"github.com/labstack/echo/v4"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
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

	return handleAuthLogin(c, s.AuthService, s.JWTService, params)
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

	return handleAuthRegister(c, s.AuthService, s.JWTService, params)
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

// handleAuthRegister handles the register Business Logic.
// On success, the user is created and the JWT pairs is returned.
func handleAuthRegister(c echo.Context, authService service.AuthService, jwtService service.JWTService, params RegisterParams) error {
	passwordhashed, err := util.HashPassword(params.Password)
	if err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	if err := authService.CreateAuth(c.Request().Context(), &entity.Auth{
		Source: localSource,
		User: &entity.User{
			Email:    null.StringFrom(params.Email),
			Name:     params.Name,
			Password: null.StringFrom(fmt.Sprint(passwordhashed)),
		},
	}); err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	return handleAuthLogin(c, authService, jwtService, LoginParams{
		Email:    params.Email,
		Password: params.Password,
	})
}
