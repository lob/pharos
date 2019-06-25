package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/lob/pharos/pkg/pharos-api-server/application"
	"github.com/lob/pharos/pkg/shared/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockVerifier struct{}

func (m *mockVerifier) Verify(t string) (*token.Identity, error) {
	return &token.Identity{}, nil
}

func TestNew(t *testing.T) {
	app, err := application.New()
	assert.NoError(t, err)
	app.TokenVerifier = &mockVerifier{}

	srv := New(app)

	t.Run("serves registered endpoint", func(tt *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/health", nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer pharos-v1.mockToken")
		require.Nil(t, err, "unexpected error when making new request")

		srv.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "incorrect status code")
		assert.Equal(t, `{"healthy":true}`, w.Body.String(), "incorrect response")
	})

	t.Run("handles requests for non-registered endpoints", func(tt *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/invalid/url/endpoint", nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer pharos-v1.mockToken")
		require.Nil(t, err, "unexpected error when making new request")

		srv.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "incorrect status code")
		assert.Contains(t, w.Body.String(), "Not Found", "incorrect response")
	})
}
