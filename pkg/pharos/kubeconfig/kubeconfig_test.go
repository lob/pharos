package kubeconfig

import (
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/lob/pharos/pkg/pharos/api"
	configpkg "github.com/lob/pharos/pkg/pharos/config"
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
		assert.NoError(tt, err)
		assert.Equal(tt, "sandbox", cluster)
	})

	t.Run("errors when reading from malformed config file", func(tt *testing.T) {
		_, err := CurrentCluster(malformedConfig)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "illegal base64 data at input byte 1")
	})

	t.Run("errors when reading from empty config file", func(tt *testing.T) {
		_, err := CurrentCluster(emptyConfig)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "context not found")
	})

	t.Run("errors when reading from nonexistent config file", func(tt *testing.T) {
		_, err := CurrentCluster(nonExistentConfig)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "no such file or directory")
	})
}

func TestGetCluster(t *testing.T) {
	// Set up dummy server for testing.
	var getResponse = []byte(`{
		"id":                     "sandbox-222222",
		"environment":            "sandbox",
		"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
		"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
		"object":                 "cluster",
		"active":                 false
	}`)
	var listResponse = []byte(`[{
		"id":                     "sandbox-333333",
		"environment":            "sandbox",
		"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
		"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
		"object":                 "cluster",
		"active":                 true
	}]`)
	var listResponse0 = []byte(`[]`)
	var listResponse2 = []byte(`[{},{}]`)

	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		switch r.URL.String() {
		case "/clusters/sandbox-222222":
			_, err := rw.Write(getResponse)
			require.NoError(t, err)
		case "/clusters?active=true&environment=sandbox":
			_, err := rw.Write(listResponse)
			require.NoError(t, err)
		case "/clusters?active=true&environment=test0clusters":
			_, err := rw.Write(listResponse0)
			require.NoError(t, err)
		case "/clusters?active=true&environment=test2clusters":
			_, err := rw.Write(listResponse2)
			require.NoError(t, err)
		}
	}))
	defer srv.Close()

	// Set BaseURL in config to be the url of the dummy server.
	client := api.NewClient(&configpkg.Config{BaseURL: srv.URL})

	t.Run("successfully merges new kubeconfig file from cluster", func(tt *testing.T) {
		// Create temporary test config file and defer cleanup.
		configFile := test.CopyTestFile(tt, "../testdata", "get", config)
		defer os.Remove(configFile)

		// Merge cluster information from active cluster for sandbox into configFile.
		err := GetCluster("sandbox", configFile, false, client)
		assert.NoError(tt, err)

		// Load kubeconfig file for testing.
		kubeConfig, err := configFromFile(configFile)
		assert.NoError(tt, err)

		// Load old kubeconfig file for comparison.
		oldKubeConfig, err := configFromFile(config)
		assert.NoError(tt, err)

		// Check that context for sandbox has been updated.
		context, ok := kubeConfig.Contexts["sandbox"]
		assert.True(tt, ok)
		assert.Equal(tt, context.Cluster, "sandbox-333333")
		assert.Equal(tt, context.AuthInfo, "engineering-sandbox-333333")

		// Check that context for the new cluster exists in the file.
		context, ok = kubeConfig.Contexts["sandbox-333333"]
		assert.True(tt, ok)
		assert.Equal(tt, context.Cluster, "sandbox-333333")
		assert.Equal(tt, context.AuthInfo, "engineering-sandbox-333333")

		// Check that new user was created for new cluster.
		user, ok := kubeConfig.AuthInfos["engineering-sandbox-333333"]
		assert.True(tt, ok)
		assert.Equal(tt, user.Exec.Command, "aws-iam-authenticator")
		assert.Equal(tt, user.Exec.Args, []string{"token", "-i", "sandbox-333333"})

		// Check that current context has not been modified.
		assert.Equal(tt, oldKubeConfig.CurrentContext, kubeConfig.CurrentContext)
	})

	t.Run("successfully creates new kubeconfig file when no previous kubeconfig file exists", func(tt *testing.T) {
		nonExistentConfig := "../testdata/nonexistentFile"
		defer os.Remove(nonExistentConfig)

		// Merge cluster information from active cluster for sandbox into nonexistent file.
		err := GetCluster("sandbox", nonExistentConfig, false, client)
		assert.NoError(tt, err)

		// Load kubeconfig file for testing.
		kubeConfig, err := configFromFile(nonExistentConfig)
		assert.NoError(tt, err)

		// Check that current context has been set because we are starting from a nonexistent file.
		assert.Equal(tt, "sandbox", kubeConfig.CurrentContext)
	})

	t.Run("successfully creates new kubeconfig file when kubeconfig file is empty", func(tt *testing.T) {
		// Create temporary test config file and defer cleanup.
		configFile := test.CopyTestFile(tt, "../testdata", "get", emptyConfig)
		defer os.Remove(configFile)

		// Merge cluster information from active cluster for sandbox into configFile.
		err := GetCluster("sandbox-222222", configFile, false, client)
		assert.NoError(tt, err)

		// Load kubeconfig file for testing.
		kubeConfig, err := configFromFile(configFile)
		assert.NoError(tt, err)

		// Check that current context has been set because we are starting from an empty file.
		assert.Equal(tt, "sandbox-222222", kubeConfig.CurrentContext)
	})

	t.Run("takes no action when --dry-run flag is set", func(tt *testing.T) {
		oldKubeConfig, err := configFromFile(config)
		assert.NoError(tt, err)

		// Run get cluster with dry-run.
		err = GetCluster("sandbox", config, true, client)
		assert.NoError(tt, err)

		// Check that kubeconfig file has not been modified.
		kubeConfig, err := configFromFile(config)
		assert.NoError(tt, err)
		assert.True(tt, reflect.DeepEqual(oldKubeConfig, kubeConfig))
	})

	t.Run("errors on merging with malformed kubeconfig file", func(tt *testing.T) {
		err := GetCluster("sandbox", malformedConfig, true, client)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "unable to load kubeconfig file")
	})

	t.Run("errors related to retrieving cluster information from the pharos API", func(tt *testing.T) {
		// Failed to list cluster.
		err := GetCluster("production", config, false, client)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "unable to list clusters for specified environment")

		// Failed to get cluster.
		err = GetCluster("sandbox-707070", config, false, client)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "failed to get cluster")

		// Received zero clusters from list cluster.
		err = GetCluster("test0clusters", config, true, client)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "no active cluster found for environment")

		// Received too many clusters from list cluster.
		err = GetCluster("test2clusters", config, true, client)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "multiple clusters found for environment")
	})
}

func TestSwitchCluster(t *testing.T) {
	t.Run("successfully switches to cluster", func(tt *testing.T) {
		// Create temporary test config file and defer cleanup.
		configFile := test.CopyTestFile(tt, "../testdata", "switch", config)
		defer os.Remove(configFile)

		// Check that current cluster is "sandbox".
		cluster, err := CurrentCluster(configFile)
		assert.NoError(tt, err)
		assert.Equal(tt, "sandbox", cluster)

		// Switch to context "sandbox-111111".
		err = SwitchCluster(configFile, "sandbox-111111")
		assert.NoError(tt, err)

		// Check that switch was successful.
		cluster, err = CurrentCluster(configFile)
		assert.NoError(tt, err)
		assert.Equal(tt, "sandbox-111111", cluster)
	})

	t.Run("errors when switching to a cluster that does not exist", func(tt *testing.T) {
		cluster, err := CurrentCluster(config)
		assert.NoError(tt, err)
		assert.Equal(tt, "sandbox", cluster)

		// Switch to context "egg".
		err = SwitchCluster(config, "egg")
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "cluster does not exist in context")

		// Current cluster should still be set to sandbox.
		cluster, err = CurrentCluster(config)
		assert.NoError(tt, err)
		assert.Equal(tt, "sandbox", cluster)
	})

	t.Run("errors when switching using malformed config file", func(tt *testing.T) {
		err := SwitchCluster(malformedConfig, "sandbox")
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "illegal base64 data at input byte 1")
	})

	t.Run("errors when switching using empty config file", func(tt *testing.T) {
		err := SwitchCluster(emptyConfig, "sandbox")
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "cluster does not exist in context")
	})

	t.Run("errors when switching using nonexistent config file", func(tt *testing.T) {
		err := SwitchCluster(nonExistentConfig, "sandbox")
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "no such file or directory")
	})
}

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
