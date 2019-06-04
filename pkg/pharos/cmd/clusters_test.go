package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const Config string = "../testdata/config"
const MalformedConfig string = "../testdata/malformed"
const NonExistentConfig string = "../testdata/nonexistent"
const EmptyConfig string = "../testdata/empty"

func TestRunClusters(t *testing.T) {
	t.Run("successfully retrieves current cluster", func(tt *testing.T) {
		err := runClusters(Config)
		assert.NoError(t, err)
	})

	t.Run("errors when reading from malformed config file", func(tt *testing.T) {
		err := runClusters(MalformedConfig)
		assert.Error(t, err)
	})

	t.Run("errors when reading from nonexistent config file", func(tt *testing.T) {
		err := runClusters(NonExistentConfig)
		assert.Error(t, err)
	})

	t.Run("errors when reading from empty config file", func(tt *testing.T) {
		err := runClusters(EmptyConfig)
		assert.Error(t, err)
	})
}
