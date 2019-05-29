package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthRoute(t *testing.T) {
	e := echo.New()

	RegisterRoutes(e)

	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "incorrect status code")
	assert.Equal(t, `{"healthy":true}`, w.Body.String(), "incorrect response")
}
