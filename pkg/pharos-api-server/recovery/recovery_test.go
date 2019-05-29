package recovery

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecovery(t *testing.T) {
	e := echo.New()
	e.Use(Middleware())

	e.GET("/panic", func(c echo.Context) error {
		panic(errors.New("panic test"))
	})

	req, err := http.NewRequest("GET", "/panic", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code, "incorrect recovered status code")
	assert.Contains(t, w.Body.String(), "Internal Server Error", "incorrect error message")
}
