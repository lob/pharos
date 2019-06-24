package cli

import (
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/lob/pharos/internal/test"
	"github.com/lob/pharos/pkg/pharos/api"
	configpkg "github.com/lob/pharos/pkg/pharos/config"
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
	getResponse := []byte(`{
		"id":                     "sandbox-222222",
		"environment":            "sandbox",
		"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
		"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
		"object":                 "cluster",
		"active":                 false
	}`)
	listResponse := []byte(`[{
		"id":                     "sandbox-333333",
		"environment":            "sandbox",
		"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
		"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
		"object":                 "cluster",
		"active":                 true
	}]`)
	listResponse0 := []byte(`[]`)
	listResponse2 := []byte(`[{},{}]`)
	listResponse3 := []byte(`[{
		"id":                     "platform-postmasters-777777",
		"environment":            "platform-postmasters",
		"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
		"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
		"object":                 "cluster",
		"active":                 true
	}]`)

	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		var response []byte
		switch r.URL.String() {
		case "/clusters/sandbox-222222":
			response = getResponse
		case "/clusters?active=true&environment=sandbox":
			response = listResponse
		case "/clusters?active=true&environment=test0clusters":
			response = listResponse0
		case "/clusters?active=true&environment=test2clusters":
			response = listResponse2
		case "/clusters?active=true&environment=platform-postmasters":
			response = listResponse3
		}
		_, err := rw.Write(response)
		require.NoError(t, err)
	}))
	defer srv.Close()
	tokenGenerator := test.NewGenerator()

	// Set BaseURL in config to be the url of the dummy server.
	client := api.NewClient(&configpkg.Config{BaseURL: srv.URL}, tokenGenerator)

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
		assert.Equal(tt, "sandbox-333333", context.Cluster)
		assert.Equal(tt, "iam-sandbox-333333", context.AuthInfo)

		// Check that context for the new cluster exists in the file.
		context, ok = kubeConfig.Contexts["sandbox-333333"]
		assert.True(tt, ok)
		assert.Equal(tt, "sandbox-333333", context.Cluster)
		assert.Equal(tt, "iam-sandbox-333333", context.AuthInfo)

		// Check that new user was created for new cluster.
		user, ok := kubeConfig.AuthInfos["iam-sandbox-333333"]
		assert.True(tt, ok)
		assert.Equal(tt, "aws-iam-authenticator", user.Exec.Command)
		assert.Equal(tt, []string{"token", "-i", "sandbox-333333"}, user.Exec.Args)

		// Check that current context has not been modified.
		assert.Equal(tt, kubeConfig.CurrentContext, oldKubeConfig.CurrentContext)
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
		assert.Contains(tt, err.Error(), "2 clusters found for environment")
	})

	t.Run("successfully merges new kubeconfig file from cluster using environment with more than one dash into an empty file", func(tt *testing.T) {
		// Create temporary test config file and defer cleanup.
		configFile := test.CopyTestFile(tt, "../testdata", "get", emptyConfig)
		defer os.Remove(configFile)

		// Merge cluster information from active cluster for sandbox into configFile.
		err := GetCluster("platform-postmasters", configFile, false, client)
		assert.NoError(tt, err)

		// Load kubeconfig file for testing.
		kubeConfig, err := configFromFile(configFile)
		assert.NoError(tt, err)

		// Check that context has been updated.
		context, ok := kubeConfig.Contexts["platform-postmasters"]
		assert.True(tt, ok)
		assert.Equal(tt, "platform-postmasters-777777", context.Cluster)
		assert.Equal(tt, "iam-platform-postmasters-777777", context.AuthInfo)

		// Check that new user was created for new cluster.
		user, ok := kubeConfig.AuthInfos["iam-platform-postmasters-777777"]
		assert.True(tt, ok)
		assert.Equal(tt, []string{"token", "-i", "platform-postmasters-777777"}, user.Exec.Args)
		assert.Equal(tt, clientcmdapi.ExecEnvVar{Name: "AWS_PROFILE", Value: "platform-postmasters"}, user.Exec.Env[0])

		// Check that current context has been set.
		assert.Equal(tt, "platform-postmasters", kubeConfig.CurrentContext)
	})
}

func TestListClusters(t *testing.T) {
	// Set up dummy server for testing.
	listClusters := []byte(`[{
		"id":                     "production-eggs",
		"environment":            "production",
		"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
		"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
		"object":                 "cluster",
		"active":                 false
	},{
		"id":                     "sandbox-333333",
		"environment":            "sandbox",
		"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
		"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
		"object":                 "cluster",
		"active":                 true
	},{
		"id":                     "sandbox-222222",
		"environment":            "sandbox",
		"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
		"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
		"object":                 "cluster",
		"active":                 false
	}]`)
	listSandbox := []byte(`[{
		"id":                     "sandbox-333333",
		"environment":            "sandbox",
		"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
		"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
		"object":                 "cluster",
		"active":                 true
	},{
		"id":                     "sandbox-222222",
		"environment":            "sandbox",
		"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
		"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
		"object":                 "cluster",
		"active":                 false
	}]`)
	listActiveStaging := []byte(`[{
		"id":                     "staging-555555",
		"environment":            "staging",
		"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
		"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
		"object":                 "cluster",
		"active":                 true
	}]`)

	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		var response []byte
		switch r.URL.String() {
		case "/clusters":
			response = listClusters
		case "/clusters?environment=sandbox":
			response = listSandbox
		case "/clusters?active=true&environment=staging":
			response = listActiveStaging
		}
		_, err := rw.Write(response)
		require.NoError(t, err)
	}))
	defer srv.Close()
	tokenGenerator := test.NewGenerator()

	// Set BaseURL in config to be the url of the dummy server.
	client := api.NewClient(&configpkg.Config{BaseURL: srv.URL}, tokenGenerator)

	t.Run("successfully lists all clusters", func(tt *testing.T) {
		// Lists all non-deleted clusters.
		clusters, err := ListClusters("", true, client)
		assert.NoError(tt, err)
		assert.Contains(tt, clusters, "sandbox-222222")
		assert.Contains(tt, clusters, "sandbox-333333")
		assert.Contains(tt, clusters, "production-eggs")
	})

	t.Run("successfully lists all clusters for an environment", func(tt *testing.T) {
		// List all clusters for a certain environment.
		clusters, err := ListClusters("sandbox", true, client)
		assert.NoError(tt, err)
		assert.Contains(tt, clusters, "sandbox-222222")
		assert.Contains(tt, clusters, "sandbox-333333")
	})

	t.Run("successfully lists all active clusters for an environment", func(tt *testing.T) {
		// List all active clusters for a certain environment.
		clusters, err := ListClusters("staging", false, client)
		assert.NoError(tt, err)
		assert.Contains(tt, clusters, "staging-555555")
	})

	t.Run("errors related to retrieving cluster information from the pharos API", func(tt *testing.T) {
		// Failed to list cluster.
		_, err := ListClusters("random", true, client)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "failed to list clusters")
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

func TestSyncClusters(t *testing.T) {
	// Set up dummy server for testing.
	syncResponse := []byte(`[{
		"id":                     "sandbox-333333",
		"environment":            "sandbox",
		"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
		"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
		"object":                 "cluster",
		"active":                 false
	}, {
		"id":                     "sandbox-444444",
		"environment":            "sandbox",
		"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
		"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
		"object":                 "cluster",
		"active":                 true
	}, {
		"id":                     "core-111111",
		"environment":            "core",
		"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
		"server_url":             "https://test.com",
		"object":                 "cluster",
		"active":                 true
	}, {
		"id":                     "staging-555555",
		"environment":            "staging",
		"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
		"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
		"object":                 "cluster",
		"active":                 false
	}, {
		"id":                     "staging-666666",
		"environment":            "staging",
		"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
		"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
		"object":                 "cluster",
		"active":                 true
	}]`)

	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_, err := rw.Write(syncResponse)
		require.NoError(t, err)
	}))
	defer srv.Close()
	tokenGenerator := test.NewGenerator()

	// Set BaseURL in config to be the url of the dummy server.
	client := api.NewClient(&configpkg.Config{BaseURL: srv.URL}, tokenGenerator)

	t.Run("successfully merges kubeconfig file from syncing clusters into existing kubeconfig", func(tt *testing.T) {
		// Create temporary test config file and defer cleanup.
		configFile := test.CopyTestFile(tt, "../testdata", "sync", config)
		defer os.Remove(configFile)

		// Merge cluster information from active cluster for sandbox into configFile.
		err := SyncClusters(configFile, false, false, client)
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
		assert.Equal(tt, context.Cluster, "sandbox-444444")
		assert.Equal(tt, context.AuthInfo, "iam-sandbox-444444")

		// Check that context for the sandbox cluster exists in the file.
		context, ok = kubeConfig.Contexts["sandbox-333333"]
		assert.True(tt, ok)
		assert.Equal(tt, context.Cluster, "sandbox-333333")
		assert.Equal(tt, context.AuthInfo, "iam-sandbox-333333")

		// Check that contexts for other clusters were added.
		_, ok = kubeConfig.Contexts["core"]
		assert.True(tt, ok)
		_, ok = kubeConfig.Contexts["core-111111"]
		assert.True(tt, ok)
		context, ok = kubeConfig.Contexts["staging"]
		assert.True(tt, ok)
		assert.Equal(tt, context.Cluster, "staging-666666")
		assert.Equal(tt, context.AuthInfo, "iam-staging-666666")
		_, ok = kubeConfig.Contexts["staging-555555"]
		assert.True(tt, ok)

		// Check that new users were created for all new clusters.
		_, ok = kubeConfig.AuthInfos["iam-sandbox-333333"]
		assert.True(tt, ok)
		_, ok = kubeConfig.AuthInfos["iam-sandbox-444444"]
		assert.True(tt, ok)
		_, ok = kubeConfig.AuthInfos["iam-core-111111"]
		assert.True(tt, ok)
		_, ok = kubeConfig.AuthInfos["iam-staging-555555"]
		assert.True(tt, ok)

		// Check that old clusters and contexts still exist in file.
		_, ok = kubeConfig.Clusters["sandbox-111111"]
		assert.True(tt, ok)
		_, ok = kubeConfig.Contexts["sandbox-111111"]
		assert.True(tt, ok)

		// Check that current context has not been modified.
		assert.Equal(tt, kubeConfig.CurrentContext, oldKubeConfig.CurrentContext)
	})

	t.Run("successfully overwrites existing kubeconfig using when --overwrite flag is set to true", func(tt *testing.T) {
		// Create temporary test config file and defer cleanup.
		configFile := test.CopyTestFile(tt, "../testdata", "sync", config)
		defer os.Remove(configFile)

		// Merge cluster information from active cluster for sandbox into configFile.
		err := SyncClusters(configFile, false, true, client)
		assert.NoError(tt, err)

		// Load kubeconfig file for testing.
		kubeConfig, err := configFromFile(configFile)
		assert.NoError(tt, err)

		// Check that context for sandbox has been updated.
		context, ok := kubeConfig.Contexts["sandbox"]
		assert.True(tt, ok)
		assert.Equal(tt, context.Cluster, "sandbox-444444")
		assert.Equal(tt, context.AuthInfo, "iam-sandbox-444444")

		// Check that context and clusters for old clusters no longer exists in the file.
		_, ok = kubeConfig.Clusters["sandbox-111111"]
		assert.False(tt, ok)
		_, ok = kubeConfig.Contexts["sandbox-111111"]
		assert.False(tt, ok)
	})

	t.Run("successfully creates new kubeconfig file from syncing clusters", func(tt *testing.T) {
		// Create temporary test config file and defer cleanup.
		nonExistentConfig := "../testdata/nonexistentFile"
		defer os.Remove(nonExistentConfig)

		// Merge cluster information from active cluster for sandbox into configFile.
		err := SyncClusters(nonExistentConfig, false, false, client)
		assert.NoError(tt, err)

		// Load kubeconfig file for testing.
		kubeConfig, err := configFromFile(nonExistentConfig)
		assert.NoError(tt, err)

		// Check that contexts for clusters were added.
		_, ok := kubeConfig.Contexts["core"]
		assert.True(tt, ok)
		_, ok = kubeConfig.Contexts["core-111111"]
		assert.True(tt, ok)
		context, ok := kubeConfig.Contexts["staging"]
		assert.True(tt, ok)
		assert.Equal(tt, context.Cluster, "staging-666666")
		assert.Equal(tt, context.AuthInfo, "iam-staging-666666")
		_, ok = kubeConfig.Contexts["staging-555555"]
		assert.True(tt, ok)

		// Check that new users were created for all new clusters.
		_, ok = kubeConfig.AuthInfos["iam-sandbox-333333"]
		assert.True(tt, ok)
		_, ok = kubeConfig.AuthInfos["iam-sandbox-444444"]
		assert.True(tt, ok)
		_, ok = kubeConfig.AuthInfos["iam-core-111111"]
		assert.True(tt, ok)
		_, ok = kubeConfig.AuthInfos["iam-staging-666666"]
		assert.True(tt, ok)
	})

	t.Run("takes no action when --dry-run flag is set", func(tt *testing.T) {
		oldKubeConfig, err := configFromFile(config)
		assert.NoError(tt, err)

		// Run get cluster with dry-run.
		err = SyncClusters(config, true, false, client)
		assert.NoError(tt, err)

		// Check that kubeconfig file has not been modified.
		kubeConfig, err := configFromFile(config)
		assert.NoError(tt, err)
		assert.True(tt, reflect.DeepEqual(oldKubeConfig, kubeConfig))
	})

	t.Run("errors on merging with malformed kubeconfig file", func(tt *testing.T) {
		err := SyncClusters(malformedConfig, true, false, client)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "unable to load kubeconfig file")
	})

	t.Run("errors related to retrieving cluster information from the pharos API", func(tt *testing.T) {
		// Failed to list cluster.
		err := SyncClusters(config, false, false, api.NewClient(&configpkg.Config{BaseURL: ""}, tokenGenerator))
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "failed to list clusters")
	})
}
