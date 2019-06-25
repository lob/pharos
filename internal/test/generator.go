package test

import (
	"testing"

	"github.com/lob/pharos/pkg/util/token"
)

// mockGenerator is used to mock out token generators.
type mockGenerator struct {
	s string
}

func (m mockGenerator) GetSTSToken() (string, error) {
	return m.s, nil
}

// NewGenerator returns a mock token generator.
func NewGenerator(t *testing.T) token.Generator {
	return &mockGenerator{s: "test"}
}
