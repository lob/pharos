package authentication

import (
	"strings"

	"github.com/labstack/echo"
	"github.com/lob/pharos/pkg/util/token"
)

const authPrefix = "Bearer "

type authMiddleware struct {
	tokenValidator token.Verifier
}

func extractAuthToken(c echo.Context) (string, error) {
	auth := c.Request().Header.Get(echo.HeaderAuthorization)
	if auth == "" {
		return "", echo.NewHTTPError(400, "missing authentication header")
	}

	if strings.HasPrefix(auth, authPrefix) {
		return strings.TrimPrefix(auth, authPrefix), nil
	}

	return "", echo.NewHTTPError(400, "invalid authentication scheme")
}

// Middleware attaches an authentication middleware that authenticates a request
// and attaches a token.Identity struct to the request if properly authenticated.
func Middleware(verifier token.Verifier) echo.MiddlewareFunc {
	am := authMiddleware{verifier}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key, err := extractAuthToken(c)
			if err != nil {
				return err
			}

			identity, err := am.tokenValidator.Verify(key)
			if err != nil {
				return echo.NewHTTPError(401)
			}

			c.Set("auth", identity)

			return next(c)
		}
	}
}
