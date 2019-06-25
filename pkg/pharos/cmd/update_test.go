package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lob/pharos/internal/test"
	"github.com/lob/pharos/pkg/pharos/api"
	configpkg "github.com/lob/pharos/pkg/pharos/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunUpdate(t *testing.T) {
	t.Run("successfully updates a cluster", func(tt *testing.T) {
		// Set up dummy server for testing.
		updateClusters := []byte(`{
			"id":                     "sandbox-333333",
			"environment":            "sandbox",
			"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
			"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
			"object":                 "cluster",
			"deleted":                true,
			"active":                 false
		}`)

		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			_, err := rw.Write(updateClusters)
			require.NoError(tt, err)
		}))
		defer srv.Close()
		tokenGenerator := test.NewGenerator(t)

		// Set BaseURL in config to be the url of the dummy server.
		client := api.NewClient(&configpkg.Config{BaseURL: srv.URL}, tokenGenerator)

		err := runUpdate("sandbox-333333", false, client)
		assert.NoError(tt, err)
	})

	t.Run("errors when the api server fails to respond", func(tt *testing.T) {
		// Set up dummy server for testing.
		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			_, err := rw.Write([]byte(``))
			require.NoError(tt, err)
		}))
		defer srv.Close()
		tokenGenerator := test.NewGenerator(t)

		// Set BaseURL in config to be the url of the dummy server.
		client := api.NewClient(&configpkg.Config{BaseURL: srv.URL}, tokenGenerator)

		err := runUpdate("sandbox-333333", false, client)
		assert.Error(tt, err)
	})
}
