package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cfg := New()
	assert.Equal(t, 7654, cfg.Port)
	assert.NotNil(t, cfg, "returned config shouldn't be nil")
}
