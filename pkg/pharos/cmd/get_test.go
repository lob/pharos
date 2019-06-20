package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/lob/pharos/pkg/pharos/api"
	configpkg "github.com/lob/pharos/pkg/pharos/config"
	"github.com/lob/pharos/pkg/pharos/kubeconfig"
	"github.com/lob/pharos/pkg/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunGet(t *testing.T) {
	t.Run("successfully merges information from a cluster into a kubeconfig file", func(tt *testing.T) {
		// Set up dummy server for testing.
		var testResponse = []byte(`[{
			"id": "sandbox-161616",
			"environment": "sandbox",
			"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
			"server_url": "https://test.elb.us-west-2.amazonaws.com:6443",
			"object": "cluster",
			"active": false
		}]`)
		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			_, err := rw.Write(testResponse)
			require.NoError(tt, err)
		}))
		defer srv.Close()

		// Set BaseURL in config to be the url of the dummy server.
		client := api.NewClient(&configpkg.Config{BaseURL: srv.URL})

		// Create temporary test config file and defer cleanup.
		configFile := test.CopyTestFile(tt, "../testdata", "get", config)
		defer os.Remove(configFile)

		// Merge cluster information from active cluster for sandbox into configFile.
		err := runGet("sandbox", configFile, false, client)
		assert.NoError(tt, err)

		// Check that current context has not been modified.
		clusterName, err := kubeconfig.CurrentCluster(configFile)
		assert.NoError(tt, err)
		assert.Equal(tt, "sandbox", clusterName)

		// Check that a new cluster was added by switching to it and checking whether the switch was successful.
		err = runSwitch(configFile, "sandbox-161616")
		assert.NoError(tt, err)
		clusterName, err = kubeconfig.CurrentCluster(configFile)
		assert.NoError(tt, err)
		assert.Equal(tt, "sandbox-161616", clusterName)
	})

	t.Run("errors when the api server fails to respond with a cluster", func(tt *testing.T) {
		// Set up dummy server for testing.
		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			_, err := rw.Write([]byte(`{}`))
			require.NoError(tt, err)
		}))
		defer srv.Close()

		// Set BaseURL in config to be the url of the dummy server.
		client := api.NewClient(&configpkg.Config{BaseURL: srv.URL})

		// Attempt to merge new cluster into configFile but this should fail because no cluster has been returned.
		err := runGet("sandbox", config, false, client)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "failed to get cluster information")
	})
}
