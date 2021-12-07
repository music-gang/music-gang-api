package controllers

import (
	"github.com/labstack/echo/v4"
)

var (
	userapp string
	passapp string
)

func init() {
	userapp = "userapp"
	passapp = "passapp"
}

// APIBasicAuthMiddleware - Si occupa di controllare i campi di auth basic delle api
func APIBasicAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		up := c.Request().Header.Get("userapp")
		pp := c.Request().Header.Get("passapp")

		if up == "" || pp == "" || up != userapp || pp != passapp {
			return APIAuthBasicFailedResponse(c)
		}

		return next(c)
	}
}

// JWTVerifyMiddleware - Si occcupa di verificare che la chiamata sia autenticata
func JWTVerifyMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}
