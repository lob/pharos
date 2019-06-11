package kubeconfig

import (
	"os"
	"reflect"
	"testing"

	"github.com/lob/pharos/pkg/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	config            = "../testdata/config"
	malformedConfig   = "../testdata/malformed"
	emptyConfig       = "../testdata/empty"
	nonExistentConfig = "../testdata/nonexistent"
)

func TestCurrentCluster(t *testing.T) {
	t.Run("successfully retrieves current cluster", func(tt *testing.T) {
		cluster, err := CurrentCluster(config)
		require.NoError(tt, err)
		assert.Equal(tt, "sandbox", cluster)
	})

	t.Run("errors when reading from malformed config file", func(tt *testing.T) {
		_, err := CurrentCluster(malformedConfig)
		require.Error(tt, err)
		assert.Contains(tt, err.Error(), "illegal base64 data at input byte 1")
	})

	t.Run("errors when reading from empty config file", func(tt *testing.T) {
		_, err := CurrentCluster(emptyConfig)
		require.Error(tt, err)
		assert.Contains(tt, err.Error(), "context not found")
	})

	t.Run("errors when reading from nonexistent config file", func(tt *testing.T) {
		_, err := CurrentCluster(nonExistentConfig)
		require.Error(tt, err)
		assert.Contains(tt, err.Error(), "no such file or directory")
	})
}

func TestSwitchCluster(t *testing.T) {
	t.Run("successfully switches to cluster", func(tt *testing.T) {
		// Create temporary test config file and defer cleanup.
		configFile := test.CopyTestFile(tt, "../testdata", "switch", config)
		defer os.Remove(configFile)

		// Check that current cluster is "sandbox".
		cluster, err := CurrentCluster(configFile)
		require.NoError(tt, err)
		assert.Equal(tt, "sandbox", cluster)

		// Switch to context "sandbox-111111".
		err = SwitchCluster(configFile, "sandbox-111111")
		require.NoError(tt, err)

		// Check that switch was successful.
		cluster, err = CurrentCluster(configFile)
		require.NoError(tt, err)
		assert.Equal(tt, "sandbox-111111", cluster)
	})

	t.Run("errors when switching to a cluster that does not exist", func(tt *testing.T) {
		cluster, err := CurrentCluster(config)
		require.NoError(tt, err)
		assert.Equal(tt, "sandbox", cluster)

		// Switch to context "egg".
		err = SwitchCluster(config, "egg")
		require.Error(tt, err)
		assert.Contains(tt, err.Error(), "cluster does not exist in context")

		// Current cluster should still be set to sandbox.
		cluster, err = CurrentCluster(config)
		require.NoError(tt, err)
		assert.Equal(tt, "sandbox", cluster)
	})

	t.Run("errors when switching using malformed config file", func(tt *testing.T) {
		err := SwitchCluster(malformedConfig, "sandbox")
		require.Error(tt, err)
		assert.Contains(tt, err.Error(), "illegal base64 data at input byte 1")
	})

	t.Run("errors when switching using empty config file", func(tt *testing.T) {
		err := SwitchCluster(emptyConfig, "sandbox")
		require.Error(tt, err)
		assert.Contains(tt, err.Error(), "cluster does not exist in context")
	})

	t.Run("errors when switching using nonexistent config file", func(tt *testing.T) {
		err := SwitchCluster(nonExistentConfig, "sandbox")
		require.Error(tt, err)
		assert.Contains(tt, err.Error(), "no such file or directory")
	})
}

func TestConfigFromFile(t *testing.T) {
	t.Run("successfully loads from config file", func(tt *testing.T) {
		kubeConfig, err := configFromFile(config)
		require.NoError(tt, err)
		assert.NotNil(tt, kubeConfig)
	})

	t.Run("returns empty kubeconfig struct and no error when loading from empty config file", func(tt *testing.T) {
		kubeConfig, err := configFromFile(emptyConfig)
		require.NoError(tt, err)
		assert.True(tt, reflect.DeepEqual(kubeConfig, clientcmdapi.NewConfig()))
	})

	t.Run("returns nil and error when loading from nonexistent file", func(tt *testing.T) {
		kubeConfig, err := configFromFile(nonExistentConfig)
		require.Error(tt, err)
		assert.Nil(tt, kubeConfig)
	})
}
