package cli

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestConfigFromFile(t *testing.T) {
	t.Run("successfully loads from config file", func(tt *testing.T) {
		kubeConfig, err := configFromFile(config)
		assert.NoError(tt, err)
		assert.NotNil(tt, kubeConfig)
	})

	t.Run("returns empty kubeconfig struct and no error when loading from empty config file", func(tt *testing.T) {
		kubeConfig, err := configFromFile(emptyConfig)
		assert.NoError(tt, err)
		assert.True(tt, reflect.DeepEqual(kubeConfig, clientcmdapi.NewConfig()))
	})

	t.Run("returns nil and error when loading from nonexistent file", func(tt *testing.T) {
		kubeConfig, err := configFromFile(nonExistentConfig)
		assert.Error(tt, err)
		assert.Nil(tt, kubeConfig)
	})
}
