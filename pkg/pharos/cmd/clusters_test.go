package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	config          = "../testdata/config"
	malformedConfig = "../testdata/malformed"
)

func TestRunClusters(t *testing.T) {
	t.Run("successfully retrieves current cluster", func(tt *testing.T) {
		err := runClusters(config)
		assert.NoError(tt, err)
	})

	t.Run("errors successfully when retrieving from malformed config", func(tt *testing.T) {
		err := runClusters(malformedConfig)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "Unable to retrieve cluster")
	})
}
