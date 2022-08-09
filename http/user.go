package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/service"
)

// UserHandler is the handler for the /user API.
func (s *ServerAPI) UserHandler(c echo.Context) error {

	user, err := s.ServiceHandler.CurrentAuthUser(c.Request().Context())
	if err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	return SuccessResponseJSON(c, http.StatusOK, echo.Map{
		"user": user,
	})
}

// UserUpdateHandler is the handler for the /user update API.
func (s *ServerAPI) UserUpdateHandler(c echo.Context) error {

	var userParams service.UserUpdate
	if err := c.Bind(&userParams); err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid request"), nil)
	}

	authUser, err := s.ServiceHandler.CurrentAuthUser(c.Request().Context())
	if err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	updatedUser, err := s.ServiceHandler.UpdateUser(c.Request().Context(), authUser.ID, userParams)
	if err != nil {
		return ErrorResponseJSON(c, err, nil)
	}

	return SuccessResponseJSON(c, http.StatusOK, echo.Map{
		"user": updatedUser,
	})
}
