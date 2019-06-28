package clusters

import (
	"testing"

	"github.com/labstack/echo"
	"github.com/lob/pharos/pkg/pharos-api-server/application"
	"github.com/lob/pharos/pkg/pharos-api-server/config"
	"github.com/lob/pharos/pkg/util/token"
	"github.com/stretchr/testify/assert"
)

type mockVerifier struct{}

func (m *mockVerifier) Verify(t string) (*token.Identity, error) {
	return &token.Identity{}, nil
}

func TestRegisterRoutes(t *testing.T) {
	e := echo.New()
	app := application.App{
		Config:        config.New(),
		TokenVerifier: &mockVerifier{},
	}

	RegisterRoutes(e, app)

	assert.Len(t, e.Routes(), 5)
}
