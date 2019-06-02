package kubeconfig

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const Config string = "../fixtures/config"
const MalformedConfig string = "../fixtures/malformed"
const EmptyConfig string = "../fixtures/empty"
const NonExistentConfig string = "../fixtures/nonexistent"

func TestCurrentCluster(t *testing.T) {
	t.Run("successfully retrieves current cluster", func(tt *testing.T) {
		cluster, err := CurrentCluster(Config)
		assert.NoError(t, err)
		assert.Equal(t, "sandbox", cluster)
	})

	t.Run("errors when reading from malformed config file", func(tt *testing.T) {
		_, err := CurrentCluster(MalformedConfig)
		assert.Error(t, err)
	})

	t.Run("errors when reading from nonexistent config file", func(tt *testing.T) {
		_, err := CurrentCluster(NonExistentConfig)
		assert.Error(t, err)
	})

	t.Run("errors when reading from empty config file", func(tt *testing.T) {
		_, err := CurrentCluster(EmptyConfig)
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

	t.Run("returns empty kubeconfig struct when loading from nonexistent file", func(tt *testing.T) {
		kubeConfig, err := ConfigFromFile(NonExistentConfig)
		require.NoError(t, err)
		assert.True(t, reflect.DeepEqual(kubeConfig, clientcmdapi.NewConfig()))
	})

	t.Run("returns empty kubeconfig struct when loading from empty config file", func(tt *testing.T) {
		kubeConfig, err := ConfigFromFile(EmptyConfig)
		require.NoError(t, err)
		assert.True(t, reflect.DeepEqual(kubeConfig, clientcmdapi.NewConfig()))
	})
}
