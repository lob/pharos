package cmd

import (
	"os"
	"testing"

	"github.com/lob/pharos/internal/test"
	"github.com/lob/pharos/pkg/pharos/cli"
	"github.com/stretchr/testify/assert"
)

func TestRunSwitch(t *testing.T) {
	t.Run("successfully switches to existing cluster", func(tt *testing.T) {
		// Create temporary test config file and defer cleanup.
		configFile := test.CopyTestFile(tt, "../testdata", "switch", config)
		defer os.Remove(configFile)

		// Switch to a different cluster.
		err := runSwitch(configFile, "sandbox-111111")
		assert.NoError(tt, err)

		// Check that switch was successful.
		clusterName, err := cli.CurrentCluster(configFile)
		assert.NoError(tt, err)
		assert.Equal(tt, "sandbox-111111", clusterName)
	})

	t.Run("errors when switching to a cluster that does not exist", func(tt *testing.T) {
		err := runSwitch(emptyConfig, "egg")
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "cluster switch unsuccessful")
	})
}
