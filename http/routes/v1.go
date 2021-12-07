package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/music-gang/music-gang-api/http/controllers"
)

func InitRoutesV1(e *echo.Echo) {

	// MARK: /api group

	//  api group
	apiGroup := e.Group("/api")

	// MARK: /api/v1 group

	// api/v1/ group
	v1Group := apiGroup.Group("/v1")

	v1Group.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := controllers.NewCustomContext(c, controllers.RequestTypeAPI)
			return next(cc)
		}
	})

	// MARK: /api/v1/auth group

	// api/v1/auth
	authGroup := v1Group.Group("/auth")

	authGroup.Use(controllers.APIBasicAuthMiddleware)

	// api/v1/auth/login
	authGroup.POST("/login", nil)

	// api/v1/auth/logout
	authGroup.GET("/logout", nil)

	// api/v1/auth/register
	authGroup.POST("/register", nil)
}
