package kubeconfig

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const Config string = "../fixtures/config"
const MalformedConfig string = "../fixtures/malformed"
const EmptyConfig string = "../fixtures/malformed"
const NonExistentConfig string = "../fixtures/nonexistent"

func TestCurrentCluster(t *testing.T) {
	t.Run("sucessfully retrieves current cluster", func(tt *testing.T) {
		cluster, err := CurrentCluster(Config)
		assert.NoError(t, err)
		assert.Equal(t, "sandbox-test", cluster)
	})

	t.Run("errors when reading from malformed config file", func(tt *testing.T) {
		_, err := CurrentCluster(MalformedConfig)
		assert.Error(t, err)
	})
}

func TestFilePath(t *testing.T) {
	t.Run("defaults to $HOME/.kube/config when empty string is passed in", func(tt *testing.T) {
		kubeConfigFile := FilePath("")
		assert.Equal(t, os.Getenv("HOME")+"/.kube/config", kubeConfigFile)
	})

	t.Run("returns the file name that is passed in", func(tt *testing.T) {
		kubeConfigFile := FilePath(Config)
		assert.Equal(t, "../fixtures/config", kubeConfigFile)
	})
}

func TestConfigFromFile(t *testing.T) {
	t.Run("successfully loads from config file", func(tt *testing.T) {
		_, err := ConfigFromFile(Config)
		require.NoError(t, err)
	})

	t.Run("errors when loading from nonexistent file", func(tt *testing.T) {
		_, err := ConfigFromFile(NonExistentConfig)
		assert.Error(t, err)
	})

	t.Run("errors when loading from empty config file", func(tt *testing.T) {
		_, err := ConfigFromFile(EmptyConfig)
		assert.Error(t, err)
	})
}
