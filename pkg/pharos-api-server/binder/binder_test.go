package binder

import (
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/lob/pharos/internal/test"
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
		c, _ := test.NewContext(tt, echo.GET, strings.NewReader("{}"), echo.MIMEApplicationJSON)
		p := params{}
		err := b.Bind(&p, c)
		assert.Contains(t, err.Error(), "is required")
	})

	t.Run("trims whitespace", func(tt *testing.T) {
		c, _ := test.NewContext(tt, echo.GET, strings.NewReader(`{"environment": " test ", "server_url": "https://pharos.com"}`), echo.MIMEApplicationJSON)
		p := params{}
		err := b.Bind(&p, c)
		assert.NoError(t, err)
		assert.Equal(t, p.Environment, "test")
	})

	t.Run("enforces url", func(tt *testing.T) {
		c, _ := test.NewContext(tt, echo.GET, strings.NewReader(`{"environment": "test", "server_url": "foobar"}`), echo.MIMEApplicationJSON)
		p := params{}
		err := b.Bind(&p, c)
		assert.Contains(t, err.Error(), "server_url must be a valid URL")
	})
}
