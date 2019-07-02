package clusters

import (
	"github.com/labstack/echo"
	"github.com/lob/pharos/pkg/pharos-api-server/application"
	"github.com/lob/pharos/pkg/pharos-api-server/authentication"
	"github.com/lob/pharos/pkg/pharos-api-server/authorization"
)

// RegisterRoutes takes in an Echo router and registers routes onto it.
func RegisterRoutes(e *echo.Echo, app application.App) {
	h := handler{app}

	config := app.Config

	e.GET("/clusters", h.list, authentication.Middleware(app.TokenVerifier), authorization.Middleware(config.Permissions.Read))
	e.GET("/clusters/:id", h.retrieve, authentication.Middleware(app.TokenVerifier), authorization.Middleware(config.Permissions.Read))
	e.DELETE("/clusters/:id", h.delete, authentication.Middleware(app.TokenVerifier), authorization.Middleware(config.Permissions.Admin))
	e.POST("/clusters", h.create, authentication.Middleware(app.TokenVerifier), authorization.Middleware(config.Permissions.Write))
	e.POST("/clusters/:id", h.update, authentication.Middleware(app.TokenVerifier), authorization.Middleware(config.Permissions.Admin))
}
