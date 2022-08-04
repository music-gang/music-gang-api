package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *ServerAPI) VmStatsHandler(c echo.Context) error {
	if stats, err := s.ServiceHandler.StatsVM(c.Request().Context()); err != nil {
		return ErrorResponseJSON(c, err, nil)
	} else {
		return SuccessResponseJSON(c, http.StatusOK, echo.Map{
			"stats": stats,
		})
	}
}
