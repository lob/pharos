package binder

import (
	"io"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

type params struct {
	Environment string `json:"environment" mod:"trim" validate:"required"`
	ServerURL   string `json:"server_url" validate:"required,url"`
}

func TestNew(t *testing.T) {
	b := New()
	assert.NotNil(t, b)
	assert.NotNil(t, b.binder)
	assert.NotNil(t, b.conform)
	assert.NotNil(t, b.validate)
}

func TestBind(t *testing.T) {
	b := New()
	assert.NotNil(t, b)

	t.Run("enforces required values", func(tt *testing.T) {
		c := newContext(tt, echo.GET, strings.NewReader("{}"), echo.MIMEApplicationJSON)
		p := params{}
		err := b.Bind(&p, c)
		assert.Contains(t, err.Error(), "is required")
	})

	t.Run("trims whitespace", func(tt *testing.T) {
		c := newContext(tt, echo.GET, strings.NewReader(`{"environment": " test ", "server_url": "https://pharos.com"}`), echo.MIMEApplicationJSON)
		p := params{}
		err := b.Bind(&p, c)
		assert.NoError(t, err)
		assert.Equal(t, p.Environment, "test")
	})

	t.Run("enforces url", func(tt *testing.T) {
		c := newContext(tt, echo.GET, strings.NewReader(`{"environment": "test", "server_url": "foobar"}`), echo.MIMEApplicationJSON)
		p := params{}
		err := b.Bind(&p, c)
		assert.Contains(t, err.Error(), "server_url must be a valid URL")
	})
}

// newContext returns a new echo.Context to be used for binder test. We cannot use the
// test.NewContext helper function as it imports the binder package and using it creates
// a circular dependency.
func newContext(t *testing.T, method string, r io.Reader, ctype string) echo.Context {
	t.Helper()

	e := echo.New()
	req := httptest.NewRequest(method, "/", r)
	req.Header.Set(echo.HeaderContentType, ctype)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	return c
}

// newQueryContext returns a new echo.Context to be used in binder query tests.
func newQueryContext(t *testing.T, query string) echo.Context {
	t.Helper()

	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/", strings.NewReader(""))
	req.URL = &url.URL{RawQuery: query}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	return c
}
