package kubeconfig

import (
	"os"
	"reflect"
	"testing"

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
		assert.Contains(tt, err.Error(), "file not found")
	})
}

func TestFilePath(t *testing.T) {
	t.Run("defaults to $HOME/.kube/config when empty string is passed in", func(tt *testing.T) {
		kubeConfigFile := filePath("")
		assert.Equal(tt, os.Getenv("HOME")+"/.kube/config", kubeConfigFile)
	})

	t.Run("returns the file name that is passed in", func(tt *testing.T) {
		kubeConfigFile := filePath(config)
		assert.Equal(tt, "../testdata/config", kubeConfigFile)
	})
}

func TestConfigFromFile(t *testing.T) {
	t.Run("successfully loads from config file", func(tt *testing.T) {
		_, err := configFromFile(config)
		require.NoError(tt, err)
	})

	t.Run("returns empty kubeconfig struct and no error when loading from empty config file", func(tt *testing.T) {
		kubeConfig, err := configFromFile(emptyConfig)
		require.NoError(tt, err)
		assert.True(tt, reflect.DeepEqual(kubeConfig, clientcmdapi.NewConfig()))
	})

	t.Run("returns empty kubeconfig struct and error when loading from nonexistent file", func(tt *testing.T) {
		kubeConfig, err := configFromFile(nonExistentConfig)
		require.Error(tt, err)
		assert.Contains(tt, err.Error(), "file not found")
		assert.True(tt, reflect.DeepEqual(kubeConfig, clientcmdapi.NewConfig()))
	})
}
