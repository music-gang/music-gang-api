package http

import (
	"github.com/labstack/echo/v4"
	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
)

const (
	claimsContextParam = "claims"
)

// JWTVerifyMiddleware is the middleware for validating JWT tokens.
// It is used for all routes that require authentication.
// Once the token is successfully parsed, checks if user exists with the auth stored in the claims.
func (s *ServerAPI) JWTVerifyMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		claims, err := s.JWTService.Parse(c.Request().Context(), extractJWT(c.Request()))
		if err != nil {
			return ErrorResponseJSON(c, err, nil)
		}

		// load user by id and check if user is active with specific auth
		if _, err := s.UserService.FindUserByID(c.Request().Context(), claims.Auth.ID); err != nil {
			return ErrorResponseJSON(c, err, nil)
		}
		if _, err := s.AuthService.FindAuthByID(c.Request().Context(), claims.Auth.ID); err != nil {
			return ErrorResponseJSON(c, err, nil)
		}

		setClaimsInContext(c, claims)

		return next(c)
	}
}

// RecoverPanicMiddleware is the middleware for handling panics.
func (s *ServerAPI) RecoverPanicMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		defer func() {
			if r := recover(); r != nil {
				err := apperr.Errorf(apperr.EUNKNOWN, "panic: %s", r)
				ErrorResponseJSON(c, err, nil)
			}
		}()

		return next(c)
	}
}

// authUser returns the current authenticated user from the context.
// Returns EUNAUTHORIZED error if no user is found in the context.
func authUser(c echo.Context) (*entity.User, error) {

	if claims, ok := c.Get(claimsContextParam).(*entity.AppClaims); ok {
		return claims.Auth.User, nil
	}

	// this should never happen
	return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "no auth user found in context")
}

// setClaimsInContext sets the claims in the context.
func setClaimsInContext(c echo.Context, claims *entity.AppClaims) {
	c.Set(claimsContextParam, claims)
	c.SetRequest(
		c.Request().WithContext(
			app.NewContextWithUser(c.Request().Context(), claims.Auth.User),
		),
	)
}
