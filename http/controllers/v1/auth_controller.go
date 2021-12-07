package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/music-gang/music-gang-api/http/controllers"
)

func Login(c echo.Context) error {

	return c.JSON(http.StatusOK, controllers.SuccessResponse(c, nil))
}
