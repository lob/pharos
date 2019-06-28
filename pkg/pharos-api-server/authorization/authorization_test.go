package authorization

import (
	"testing"

	"github.com/labstack/echo"
	"github.com/lob/pharos/pkg/util/token"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	e := echo.New()

	t.Run("succesfully authorizes a valid request", func(tt *testing.T) {
		c := e.NewContext(nil, nil)
		c.Set("auth", &token.Identity{CanonicalARN: "admin"})

		m := Middleware([]string{"admin"})

		err := m(func(c echo.Context) error { return nil })(c)
		assert.NoError(t, err)
	})

	t.Run("rejects an invalid authorization request", func(tt *testing.T) {
		c := e.NewContext(nil, nil)
		c.Set("auth", &token.Identity{CanonicalARN: "read"})

		m := Middleware([]string{"admin"})

		err := m(func(c echo.Context) error { return nil })(c)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Unauthorized")
	})

	t.Run("rejects a request if auth is not set", func(tt *testing.T) {
		c := e.NewContext(nil, nil)

		m := Middleware([]string{"admin"})

		err := m(func(c echo.Context) error { return nil })(c)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Unauthorized")
	})

}
