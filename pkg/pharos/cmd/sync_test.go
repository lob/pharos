package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/lob/pharos/internal/test"
	"github.com/lob/pharos/pkg/pharos/api"
	"github.com/lob/pharos/pkg/pharos/cli"
	configpkg "github.com/lob/pharos/pkg/pharos/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSync(t *testing.T) {
	t.Run("successfully merges information from a cluster into a kubeconfig file", func(tt *testing.T) {
		// Set up dummy server for testing.
		testResponse := []byte(`[{
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
			_, err := rw.Write(testResponse)
			require.NoError(tt, err)
		}))
		defer srv.Close()
		tokenGenerator := test.NewGenerator()

		// Set BaseURL in config to be the url of the dummy server.
		client := api.NewClient(&configpkg.Config{BaseURL: srv.URL}, tokenGenerator)

		// Create temporary test config file and defer cleanup.
		configFile := test.CopyTestFile(tt, "../testdata", "get", config)
		defer os.Remove(configFile)

		// Merge cluster information from active cluster for sandbox into configFile.
		err := runSync(configFile, false, false, client)
		assert.NoError(tt, err)

		// Check that current context has not been modified.
		clusterName, err := cli.CurrentCluster(configFile)
		assert.NoError(tt, err)
		assert.Equal(tt, "sandbox", clusterName)

		// Check that a new context for staging was added by switching to it and checking whether the switch was successful.
		err = runSwitch(configFile, "staging")
		assert.NoError(tt, err)
		clusterName, err = cli.CurrentCluster(configFile)
		assert.NoError(tt, err)
		assert.Equal(tt, "staging", clusterName)
	})

	t.Run("errors when the api server fails to respond with a cluster", func(tt *testing.T) {
		// Set up dummy server for testing.
		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			_, err := rw.Write([]byte(`{}`))
			require.NoError(tt, err)
		}))
		defer srv.Close()
		tokenGenerator := test.NewGenerator()

		// Set BaseURL in config to be the url of the dummy server.
		client := api.NewClient(&configpkg.Config{BaseURL: srv.URL}, tokenGenerator)

		// Attempt to merge new cluster into configFile but this should fail because no cluster has been returned.
		err := runSync(config, false, false, client)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "failed to sync clusters")
	})
}
