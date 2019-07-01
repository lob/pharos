package authorization

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/lob/pharos/pkg/util/token"
)

// Middleware attaches an authorization middleware that authenticates a request
// against the authenticated user from the authentication middleware. Attaches a
// token.Identity struct to the request if properly authenticated.
func Middleware(allowedARNs []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var authARN string
			if identity, ok := c.Get("auth").(*token.Identity); ok {
				authARN = identity.CanonicalARN
			} else {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}

			for _, allowedArn := range allowedARNs {
				if allowedArn == authARN {
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusUnauthorized)
		}
	}
}
