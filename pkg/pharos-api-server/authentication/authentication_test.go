package authentication

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/lob/pharos/pkg/shared/token"
	"github.com/stretchr/testify/assert"
)

type mockVerifier struct {
	Identity *token.Identity
	Err      error
}

func (m mockVerifier) Verify(token string) (*token.Identity, error) {
	m.Identity.AccountID = token
	return m.Identity, m.Err
}

func TestMiddleware(t *testing.T) {

	t.Run("successfully authenticates request and sets auth in context", func(tt *testing.T) {
		successVerifier := mockVerifier{
			Identity: &token.Identity{
				ARN: "arn:aws:sts:success-verifier",
			},
		}
		m := Middleware(successVerifier)

		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", strings.NewReader(""))
		req.Header.Set(echo.HeaderAuthorization, "Bearer pharos-v1.pharos-success-test")
		c := e.NewContext(req, httptest.NewRecorder())

		err := m(func(c echo.Context) error {
			i := c.Get("auth").(*token.Identity)

			assert.Equal(t, "arn:aws:sts:success-verifier", i.ARN)
			assert.Equal(t, "pharos-v1.pharos-success-test", i.AccountID)

			return nil
		})(c)

		assert.NoError(t, err)
	})

	t.Run("returns an unauthorized error if token validation fails", func(tt *testing.T) {
		failedVerifier := mockVerifier{
			Identity: &token.Identity{},
			Err:      errors.New("token validation failed"),
		}
		m := Middleware(failedVerifier)

		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", strings.NewReader(""))
		req.Header.Set(echo.HeaderAuthorization, "Bearer pharos-v1.pharos-success-test")
		c := e.NewContext(req, httptest.NewRecorder())

		err := m(func(c echo.Context) error {
			return nil
		})(c)

		assert.Equal(t, err.(*echo.HTTPError).Code, 401)
		assert.Contains(t, err.Error(), "Unauthorized")
	})

	t.Run("returns an error if authentication header is missing", func(tt *testing.T) {
		m := Middleware(mockVerifier{})

		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", strings.NewReader(""))
		c := e.NewContext(req, httptest.NewRecorder())

		err := m(func(c echo.Context) error {
			return nil
		})(c)

		assert.Contains(t, err.Error(), "missing authentication header")
	})

	t.Run("returns an error if authentication header is incorrect", func(tt *testing.T) {
		m := Middleware(mockVerifier{})

		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", strings.NewReader(""))
		req.Header.Set(echo.HeaderAuthorization, "Foo bar")
		c := e.NewContext(req, httptest.NewRecorder())

		err := m(func(c echo.Context) error {
			return nil
		})(c)

		assert.Contains(t, err.Error(), "invalid authentication scheme")
	})

}
