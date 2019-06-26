package clusters

import (
	"testing"

	"github.com/labstack/echo"
	"github.com/lob/pharos/pkg/pharos-api-server/application"
	"github.com/stretchr/testify/assert"
)

func TestRegisterRoutes(t *testing.T) {
	e := echo.New()
	app := application.App{}

	RegisterRoutes(e, app)

	assert.Len(t, e.Routes(), 5)
}
