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

func TestRunCurrent(t *testing.T) {
	t.Run("successfully retrieves current cluster", func(tt *testing.T) {
		err := runCurrent(config)
		require.NoError(tt, err)
	})

	t.Run("errors successfully when retrieving from malformed config", func(tt *testing.T) {
		err := runCurrent(malformedConfig)
		require.Error(tt, err)
		assert.Contains(tt, err.Error(), "unable to retrieve cluster")
	})
}
