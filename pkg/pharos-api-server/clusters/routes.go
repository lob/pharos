package clusters

import (
	"github.com/labstack/echo"
	"github.com/lob/pharos/pkg/pharos-api-server/application"
)

// RegisterRoutes takes in an Echo router and registers routes onto it.
func RegisterRoutes(e *echo.Echo, app application.App) {
	g := e.Group("/clusters")

	h := handler{app}

	g.GET("", h.list)
	g.GET("/:id", h.retrieve)
	g.DELETE("/:id", h.delete)
	g.POST("", h.create)
}
