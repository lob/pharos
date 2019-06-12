package binder

import (
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/lob/pharos/internal/test"
	"github.com/stretchr/testify/assert"
)

type user struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestCustomBinderBind(t *testing.T) {
	cb := &customBinder{}
	u := &user{}

	t.Run("successfully binds payload", func(tt *testing.T) {
		c, _ := test.NewContext(tt, echo.GET, strings.NewReader(`{"id": 123, "name": "test"}`), echo.MIMEApplicationJSON)

		err := cb.bind(u, c)
		assert.NoError(tt, err)
		assert.Equal(tt, 123, u.ID)
		assert.Equal(tt, "test", u.Name)
	})

	t.Run("errors for unknown fields in payload", func(tt *testing.T) {
		c, _ := test.NewContext(tt, echo.GET, strings.NewReader(`{"foo": "bar"}`), echo.MIMEApplicationJSON)

		err := cb.bind(u, c)
		assert.Contains(tt, err.Error(), `unknown field "foo"`)
	})

	t.Run("errors for unknown MIME", func(tt *testing.T) {
		c, _ := test.NewContext(tt, echo.GET, strings.NewReader("test"), echo.MIMETextPlain)

		err := cb.bind(u, c)
		assert.Contains(tt, err.Error(), "Unsupported Media Type")
	})

	t.Run("errors with empty payloads for non GET or DELETE requests", func(tt *testing.T) {
		c, _ := test.NewContext(tt, echo.POST, strings.NewReader(""), echo.MIMEApplicationJSON)

		err := cb.bind(u, c)
		assert.Contains(tt, err.Error(), "request body can't be empty")
	})

	t.Run("allows empty GET request", func(tt *testing.T) {
		c, _ := test.NewContext(tt, echo.GET, strings.NewReader(""), echo.MIMEApplicationJSON)

		err := cb.bind(u, c)
		assert.NoError(tt, err)
	})

	t.Run("allows empty DELETE request", func(tt *testing.T) {
		c, _ := test.NewContext(tt, echo.DELETE, strings.NewReader(""), echo.MIMEApplicationJSON)

		err := cb.bind(u, c)
		assert.NoError(tt, err)
	})

}
