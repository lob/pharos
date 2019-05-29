package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	app, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.NotNil(t, app.Config)
	assert.NotNil(t, app.Metrics)
	assert.NotNil(t, app.Sentry)
}
