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

func TestRunCreate(t *testing.T) {
	t.Run("successfully creates a cluster", func(tt *testing.T) {
		// Set up dummy server for testing.
		createClusters := []byte(`{
			"id":                     "sandbox-333333",
			"environment":            "sandbox",
			"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
			"server_url":             "https://test.elb.us-west-2.amazonaws.com:6443",
			"object":                 "cluster",
			"deleted":                true,
			"active":                 true
		}`)

		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			_, err := rw.Write(createClusters)
			require.NoError(tt, err)
		}))
		defer srv.Close()
		tokenGenerator := test.NewGenerator()

		// Set BaseURL in config to be the url of the dummy server.
		client := api.NewClient(&configpkg.Config{BaseURL: srv.URL}, tokenGenerator)

		err := runCreate("sandbox-333333", "sandbox", "LS0tLS1CRUdJTiBDR", "https://test.elb.us-west-2.amazonaws.com:6443", client)
		assert.NoError(tt, err)
	})
}
