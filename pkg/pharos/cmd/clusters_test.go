package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	config          = "../testdata/config"
	malformedConfig = "../testdata/malformed"
)

func TestRunClusters(t *testing.T) {
	t.Run("successfully retrieves current cluster", func(tt *testing.T) {
		err := runClusters(config)
		require.NoError(tt, err)
	})

	t.Run("errors successfully when retrieving from malformed config", func(tt *testing.T) {
		err := runClusters(malformedConfig)
		require.Error(tt, err)
		assert.Contains(tt, err.Error(), "Unable to retrieve cluster")
	})
}
