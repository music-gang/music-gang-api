package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/service"
)

// UserHandler is the handler for the /user API.
func (s *ServerAPI) UserHandler(c echo.Context) error {
	return handleUser(c, s)
}

// UserUpdateHandler is the handler for the /user update API.
func (s *ServerAPI) UserUpdateHandler(c echo.Context) error {

	var userParams service.UserUpdate
	if err := c.Bind(&userParams); err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid request"), nil)
	}

	return handleUserUpdate(c, s, userParams)
}

// handleUser handles the /user business logic.
func handleUser(c echo.Context, s *ServerAPI) error {
	user, err := authUser(c)
	if err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	return SuccessResponseJSON(c, http.StatusOK, echo.Map{
		"user": user,
	})
}

// handleUserUpdate handles the /user update business logic.
func handleUserUpdate(c echo.Context, s *ServerAPI, userParams service.UserUpdate) error {

	user, err := authUser(c)
	if err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	if updatedUser, err := s.UserService.UpdateUser(c.Request().Context(), user.ID, userParams); err != nil {
		return ErrorResponseJSON(c, err, nil)
	} else {
		return SuccessResponseJSON(c, http.StatusOK, echo.Map{
			"user": updatedUser,
		})
	}
}
