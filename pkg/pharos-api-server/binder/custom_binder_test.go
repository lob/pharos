package binder

import (
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

type user struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type query struct {
	Environment string `query:"env"`
	Active      bool   `query:"active"`
	Foo         string
}

func TestCustomBinderBind(t *testing.T) {
	cb := &customBinder{}
	u := &user{}

	t.Run("successfully binds payload", func(tt *testing.T) {
		c := newContext(tt, echo.GET, strings.NewReader(`{"id": 123, "name": "test"}`), echo.MIMEApplicationJSON)

		err := cb.bind(u, c)
		assert.NoError(tt, err)
		assert.Equal(tt, 123, u.ID)
		assert.Equal(tt, "test", u.Name)
	})

	t.Run("errors for unknown fields in payload", func(tt *testing.T) {
		c := newContext(tt, echo.GET, strings.NewReader(`{"foo": "bar"}`), echo.MIMEApplicationJSON)

		err := cb.bind(u, c)
		assert.Contains(tt, err.Error(), `unknown field "foo"`)
	})

	t.Run("errors for unknown MIME", func(tt *testing.T) {
		c := newContext(tt, echo.GET, strings.NewReader("test"), echo.MIMETextPlain)

		err := cb.bind(u, c)
		assert.Contains(tt, err.Error(), "Unsupported Media Type")
	})

	t.Run("errors with empty payloads for non GET or DELETE requests", func(tt *testing.T) {
		c := newContext(tt, echo.POST, strings.NewReader(""), echo.MIMEApplicationJSON)

		err := cb.bind(u, c)
		assert.Contains(tt, err.Error(), "request body can't be empty")
	})

	t.Run("allows empty GET request", func(tt *testing.T) {
		c := newContext(tt, echo.GET, strings.NewReader(""), echo.MIMEApplicationJSON)

		err := cb.bind(u, c)
		assert.NoError(tt, err)
	})

	t.Run("allows empty DELETE request", func(tt *testing.T) {
		c := newContext(tt, echo.DELETE, strings.NewReader(""), echo.MIMEApplicationJSON)

		err := cb.bind(u, c)
		assert.NoError(tt, err)
	})

	t.Run("successfully binds query", func(tt *testing.T) {
		c := newQueryContext(tt, "env=test&active=true&Foo=bar")

		q := &query{}
		err := cb.bind(q, c)
		assert.NoError(tt, err)

		assert.Equal(tt, "test", q.Environment)
		assert.Equal(tt, true, q.Active)
		assert.Equal(tt, "bar", q.Foo)
	})

	t.Run("errors binding query with invalid fields", func(tt *testing.T) {
		c := newQueryContext(tt, "active=test")

		q := &query{}
		err := cb.bind(q, c)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "active must be a boolean")
	})

	t.Run("errors binding query with unknown fields", func(tt *testing.T) {
		c := newQueryContext(tt, "biz=baz")

		q := &query{}
		err := cb.bind(q, c)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "biz is not allowed")
	})

}
