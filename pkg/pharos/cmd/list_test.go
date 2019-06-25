package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lob/pharos/pkg/pharos/api"
	configpkg "github.com/lob/pharos/pkg/pharos/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunList(t *testing.T) {
	t.Run("successfully lists information about clusters", func(tt *testing.T) {
		// Set up dummy server for testing.
		var listSandboxClusters = []byte(`[{
			"id":                     "sandbox-333333",
			"environment":            "sandbox",
			"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
			"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
			"object":                 "cluster",
			"active":                 true
		}, {
			"id":                     "sandbox-222222",
			"environment":            "sandbox",
			"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
			"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
			"object":                 "cluster",
			"active":                 false
		}]`)

		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			_, err := rw.Write(listSandboxClusters)
			require.NoError(tt, err)
		}))
		defer srv.Close()
		// Set BaseURL in config to be the url of the dummy server.
		client := api.NewClient(&configpkg.Config{BaseURL: srv.URL})

		err := runList("sandbox", true, client)
		assert.NoError(tt, err)
	})

	t.Run("errors when the api server fails to respond with clusters", func(tt *testing.T) {
		// Set up dummy server for testing.
		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			_, err := rw.Write([]byte(`{}`))
			require.NoError(tt, err)
		}))
		defer srv.Close()

		// Set BaseURL in config to be the url of the dummy server.
		client := api.NewClient(&configpkg.Config{BaseURL: srv.URL})

		err := runList("", true, client)
		assert.Error(tt, err)
	})
}
