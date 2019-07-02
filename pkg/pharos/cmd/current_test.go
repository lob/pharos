package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunCurrent(t *testing.T) {
	t.Run("successfully retrieves current cluster", func(tt *testing.T) {
		err := runCurrent(config)
		assert.NoError(tt, err)
	})

	t.Run("errors successfully when retrieving from malformed config", func(tt *testing.T) {
		err := runCurrent(malformedConfig)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "unable to retrieve cluster")
	})
}
