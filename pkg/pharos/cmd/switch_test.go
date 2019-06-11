package cmd

import (
	"testing"

	"github.com/lob/pharos/pkg/pharos/kubeconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSwitch(t *testing.T) {
	t.Run("successfully switches to existing cluster", func(tt *testing.T) {
		// Switch back to context "sandbox".
		defer func() {
			err := runSwitch(config, "sandbox")
			require.NoError(tt, err)
		}()

		err := runSwitch(config, "sandbox-111111")
		require.NoError(tt, err)

		// Check that switch was successful.
		clusterName, err := kubeconfig.CurrentCluster(config)
		require.NoError(tt, err)
		require.Equal(tt, "sandbox-111111", clusterName)
	})

	t.Run("errors when switching to a cluster that does not exist", func(tt *testing.T) {
		err := runSwitch(emptyConfig, "egg")
		require.Error(tt, err)
		assert.Contains(tt, err.Error(), "cluster switch unsuccessful")
	})
}

func TestArgSwitch(t *testing.T) {
	t.Run("succeeds when we are given exactly one argument", func(tt *testing.T) {
		err := argSwitch([]string{"1"})
		require.NoError(tt, err)
	})

	t.Run("errors if too few arguments are passed in", func(tt *testing.T) {
		err := argSwitch([]string{})
		require.Error(tt, err)
	})

	t.Run("errors if too many arguments are passed in", func(tt *testing.T) {
		err := argSwitch([]string{"1", "2"})
		require.Error(tt, err)
	})
}
