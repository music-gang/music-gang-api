package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *ServerAPI) VmStatsHandler(c echo.Context) error {
	return handleVmStats(c, s)
}

func handleVmStats(c echo.Context, s *ServerAPI) error {

	stats, err := s.FuelMeterService.Stats(c.Request().Context())
	if err != nil {
		s.LogService.ReportError(c.Request().Context(), err)
		return ErrorResponseJSON(c, err, nil)
	}

	return SuccessResponseJSON(c, http.StatusOK, &echo.Map{
		"stats": stats,
	})
}
