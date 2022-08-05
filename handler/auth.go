package handler

import (
	"context"
	"net/mail"

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

// AuthLogin handles the login Business Logic.
func (s *ServiceHandler) AuthLogin(ctx context.Context, params LoginParams) (*entity.User, *entity.TokenPair, error) {

	if err := params.validate(); err != nil {
		return nil, nil, err
	}

	auth, err := s.VmCallableService.Auhenticate(ctx, &entity.AuthUserOptions{
		Source: &localSource,
		UserParams: &entity.UserParams{
			Email:    &params.Email,
			Password: &params.Password,
		},
	})
	if err != nil {
		if apperr.ErrorCode(err) == apperr.ENOTFOUND {
			return nil, nil, apperr.Errorf(apperr.EUNAUTHORIZED, "wrong credentials")
		}
		s.Logger.Error(apperr.ErrorLog(err))
		return nil, nil, err
	}

	if auth.User.Auths != nil {
		auth.User.Auths = nil
	}

	pair, err := s.JWTService.Exchange(ctx, auth)
	if err != nil {
		s.Logger.Error(apperr.ErrorLog(err))
		return nil, nil, err
	}

	return auth.User, pair, nil
}

// AuthLogout handles the logout Business Logic.
func (s *ServiceHandler) AuthLogout(ctx context.Context, pair *entity.TokenPair) error {

	if pair.AccessToken != "" {
		if err := s.JWTService.Invalidate(ctx, pair.AccessToken, entity.AccessTokenExpiration); err != nil {
			s.Logger.Error(apperr.ErrorLog(err))
			return err
		}
	}

	if pair.RefreshToken != "" {
		if err := s.JWTService.Invalidate(ctx, pair.RefreshToken, entity.RefreshTokenExpiration); err != nil {
			s.Logger.Error(apperr.ErrorLog(err))
			return err
		}
	}

	return nil
}

// AuthRefresh handles the refresh Business Logic.
func (s *ServiceHandler) AuthRefresh(ctx context.Context, pair *entity.TokenPair) (*entity.User, *entity.TokenPair, error) {

	if pair.RefreshToken == "" {
		return nil, nil, apperr.Errorf(apperr.EINVALID, "refresh token is required")
	}

	pair, err := s.JWTService.Refresh(ctx, pair.RefreshToken)
	if err != nil {
		s.Logger.Error(apperr.ErrorLog(err))
		return nil, nil, err
	}

	claims, err := s.JWTService.Parse(ctx, pair.AccessToken)
	if err != nil {
		s.Logger.Error(apperr.ErrorLog(err))
		return nil, pair, err
	}

	if claims.Auth.User.Auths != nil {
		claims.Auth.User.Auths = nil
	}

	return claims.Auth.User, pair, nil
}

// AuthRegister handles the register Business Logic.
// On success, the user is created and the JWT pairs is returned.
func (s *ServiceHandler) AuthRegister(ctx context.Context, params RegisterParams) (*entity.User, *entity.TokenPair, error) {

	if err := params.validate(); err != nil {
		return nil, nil, err
	}

	passwordhashed, err := util.HashPassword(params.Password)
	if err != nil {
		s.Logger.Error(apperr.ErrorLog(err))
		return nil, nil, err
	}

	if err := s.VmCallableService.CreateAuth(ctx, &entity.Auth{
		Source: localSource,
		User: &entity.User{
			Email:    null.StringFrom(params.Email),
			Name:     params.Name,
			Password: null.StringFrom(string(passwordhashed)),
		},
	}); err != nil {
		s.Logger.Error(apperr.ErrorLog(err))
		return nil, nil, err
	}

	return s.AuthLogin(ctx, LoginParams{
		Email:    params.Email,
		Password: params.Password,
	})
}
