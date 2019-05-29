package health

import (
	"net/http"

	"github.com/labstack/echo"
)

var resp = []byte(`{"healthy":true}`)

// RegisterRoutes takes in an Echo router and registers routes onto it.
func RegisterRoutes(e *echo.Echo) {
	e.GET("/health", healthHandler)
}

func healthHandler(c echo.Context) error {
	return c.JSONBlob(http.StatusOK, resp)
}
