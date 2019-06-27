package test

import (
	"github.com/lob/pharos/pkg/util/token"
)

// mockGenerator is used to mock out token generators.
type mockGenerator struct{}

func (m mockGenerator) GetSTSToken() (string, error) {
	return "test", nil
}

// NewGenerator returns a mock token generator.
func NewGenerator() token.Generator {
	return &mockGenerator{}
}
