package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lob/pharos/pkg/pharos/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteCluster(t *testing.T) {
	var testResponse = []byte(`{
		"id":                     "production-pikachu",
		"environment":            "production",
		"server_url":             "https://prod.elb.us-west-2.amazonaws.com:6443",
		"cluster_authority_data": "asdasd",
		"deleted":                true,
		"active":                 true
	}`)

	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_, err := rw.Write(testResponse)
		require.NoError(t, err)
	}))
	defer srv.Close()

	t.Run("deletes cluster by ID successfully", func(tt *testing.T) {
		c := NewClient(&config.Config{BaseURL: srv.URL})
		cluster, err := c.DeleteCluster("production-pikachu")
		assert.NoError(tt, err)
		assert.Equal(tt, "production-pikachu", cluster.ID)
		assert.Equal(tt, true, cluster.Deleted)
	})

	t.Run("fails to delete cluster using a bad client", func(tt *testing.T) {
		c := NewClient(&config.Config{BaseURL: ""})
		cluster, err := c.DeleteCluster("production-pikachu")
		assert.Error(tt, err)
		assert.Equal(tt, "", cluster.ID)
	})
}

func TestListClusters(t *testing.T) {
	var testResponse = []byte(`[
		{
			"id": "production-6906ce",
			"environment": "production",
			"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
			"server_url": "https://prod.elb.us-west-2.amazonaws.com:6443",
			"object": "cluster",
			"active": true
		},
		{
			"id": "production-111111",
			"environment": "production",
			"cluster_authority_data": "LS0tLS1CRsdJTiBDR...",
			"server_url": "https://prod.elb.us-west-2.amazonaws.com:6443",
			"object": "cluster",
			"active": false
		}
	]`)

	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_, err := rw.Write(testResponse)
		require.NoError(t, err)
	}))
	defer srv.Close()

	t.Run("lists clusters successfully", func(tt *testing.T) {
		c := NewClient(&config.Config{BaseURL: srv.URL})
		clusters, err := c.ListClusters(nil)
		assert.NoError(tt, err)

		assert.Equal(tt, 2, len(clusters))
		assert.Equal(tt, "production-6906ce", (clusters)[0].ID)
	})

	t.Run("fails to list clusters using a bad client", func(tt *testing.T) {
		c := NewClient(&config.Config{BaseURL: ""})
		clusters, err := c.ListClusters(nil)
		assert.Error(tt, err)
		assert.Nil(tt, clusters)
	})
}

func TestGetCluster(t *testing.T) {
	var testResponse = []byte(`{
		"id": "production-6906ce",
		"environment": "production",
		"cluster_authority_data": "LS0tLS1CRUdJTiBDR...",
		"server_url": "https://prod.elb.us-west-2.amazonaws.com:6443",
		"object": "cluster",
		"active": true
	}`)

	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_, err := rw.Write(testResponse)
		require.NoError(t, err)
	}))
	defer srv.Close()

	t.Run("retrieves cluster by ID successfully", func(tt *testing.T) {
		c := NewClient(&config.Config{BaseURL: srv.URL})
		cluster, err := c.GetCluster("production-6906ce")
		assert.NoError(tt, err)
		assert.Equal(tt, "production", cluster.Environment)
	})

	t.Run("fails to retrieve cluster using a bad client", func(tt *testing.T) {
		c := NewClient(&config.Config{BaseURL: ""})
		cluster, err := c.GetCluster("production-6906ce")
		assert.Error(tt, err)
		assert.Equal(tt, "", cluster.ID)
	})
}
