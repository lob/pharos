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

func TestRunDelete(t *testing.T) {
	t.Run("successfully deletes a cluster", func(tt *testing.T) {
		// Set up dummy server for testing.
		var deleteClusters = []byte(`{
			"id":                     "sandbox-333333",
			"environment":            "sandbox",
			"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
			"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
			"object":                 "cluster",
			"deleted":                true,
			"active":                 true
		}`)

		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			_, err := rw.Write(deleteClusters)
			require.NoError(tt, err)
		}))
		defer srv.Close()

		// Set BaseURL in config to be the url of the dummy server.
		client := api.NewClient(&configpkg.Config{BaseURL: srv.URL})

		err := runDelete("sandbox-333333", client)
		assert.NoError(tt, err)
	})

	t.Run("errors when attempting to delete a nonexistent cluster", func(tt *testing.T) {
		// Set up dummy server for testing.
		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(http.StatusNotFound)
			_, err := rw.Write([]byte(`{"error":{"message":"cluster not found","status_code":404}}`))
			require.NoError(tt, err)
		}))
		defer srv.Close()

		// Set BaseURL in config to be the url of the dummy server.
		client := api.NewClient(&configpkg.Config{BaseURL: srv.URL})

		err := runDelete("sandbox-egg", client)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "failed to delete cluster sandbox-egg")
		assert.Contains(tt, err.Error(), "cluster not found")
	})
}
