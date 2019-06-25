package clusters

import (
	"github.com/labstack/echo"
	"github.com/lob/pharos/pkg/pharos-api-server/application"
)

// RegisterRoutes takes in an Echo router and registers routes onto it.
func RegisterRoutes(e *echo.Echo, app application.App) {
	h := handler{app}

	e.GET("/clusters", h.list)
	e.GET("/clusters/:id", h.retrieve)
	e.DELETE("/clusters/:id", h.delete)
	e.POST("/clusters", h.create)
}
