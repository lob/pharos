package test

import (
	"io"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/labstack/echo"
	"github.com/lob/pharos/pkg/pharos-api-server/binder"
)

// NewContext returns a new echo.Context, and *httptest.ResponseRecorder to be
// used for tests.
func NewContext(t *testing.T, method string, query string, r io.Reader, ctype string) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()

	e := echo.New()
	e.Binder = binder.New()
	req := httptest.NewRequest(method, "/", r)
	req.URL = &url.URL{RawQuery: query}
	req.Header.Set(echo.HeaderContentType, ctype)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	return c, rec
}
