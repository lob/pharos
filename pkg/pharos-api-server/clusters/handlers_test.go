package clusters

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/lob/pharos/internal/test"
	"github.com/lob/pharos/pkg/pharos-api-server/application"
	"github.com/lob/pharos/pkg/util/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	defaultTestCluster = model.Cluster{
		ID:                   "test-1",
		Environment:          "test",
		ServerURL:            "http://test-1.localhost:6443",
		ClusterAuthorityData: "abcdef",
	}
	otherTestCluster = model.Cluster{
		ID:                   "test-2",
		Environment:          "test",
		ServerURL:            "http://test-2.localhost:6443",
		ClusterAuthorityData: "abcdef",
	}
	deletedTestCluster = model.Cluster{
		ID:                   "test-3",
		Environment:          "test",
		ServerURL:            "http://test-3.localhost:6443",
		ClusterAuthorityData: "abcdef",
		Deleted:              true,
	}
	activeTestCluster = model.Cluster{
		ID:                   "test-active",
		Environment:          "test",
		ServerURL:            "http://test-3.localhost:6443",
		ClusterAuthorityData: "abcdef",
		Active:               true,
	}
	differentEnvironmentCluster = model.Cluster{
		ID:                   "other-1",
		Environment:          "other",
		ServerURL:            "http://test-3.localhost:6443",
		ClusterAuthorityData: "abcdef",
	}
)

func TestListHandler(t *testing.T) {
	h := newHandler(t)

	t.Run("lists clusters ordered correctly", func(tt *testing.T) {
		test.TruncateTables(tt, h.app.DB)
		clusters := []model.Cluster{defaultTestCluster, otherTestCluster, deletedTestCluster}
		err := h.app.DB.Insert(&clusters)
		require.NoError(tt, err)

		c, rr := test.NewContext(tt, "GET", "", strings.NewReader(""), "application/json")

		err = h.list(c)
		assert.NoError(tt, err)
		assert.Equal(tt, http.StatusOK, rr.Code)

		var response []model.Cluster
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(tt, err)
		assert.Len(tt, response, 2)
		assert.Equal(tt, defaultTestCluster.ID, response[0].ID)
		assert.Equal(tt, otherTestCluster.ID, response[1].ID)
	})

	t.Run("filters lists clusters correctly", func(tt *testing.T) {
		test.TruncateTables(tt, h.app.DB)
		clusters := []model.Cluster{defaultTestCluster, otherTestCluster, deletedTestCluster, activeTestCluster, differentEnvironmentCluster}
		err := h.app.DB.Insert(&clusters)
		require.NoError(tt, err)

		c, rr := test.NewContext(tt, "GET", "environment=test&active=true", strings.NewReader(""), "application/json")

		err = h.list(c)
		assert.NoError(tt, err)
		assert.Equal(tt, http.StatusOK, rr.Code)

		var response []model.Cluster
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(tt, err)
		assert.Len(tt, response, 1)
		assert.Equal(tt, activeTestCluster.ID, response[0].ID)
	})

	t.Run("does not list deleted clusters", func(tt *testing.T) {
		test.TruncateTables(tt, h.app.DB)
		clusters := []model.Cluster{deletedTestCluster}
		err := h.app.DB.Insert(&clusters)
		require.NoError(tt, err)

		c, rr := test.NewContext(tt, "GET", "", strings.NewReader(""), "application/json")

		err = h.list(c)
		assert.NoError(tt, err)
		assert.Equal(tt, http.StatusOK, rr.Code)

		var response []model.Cluster
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(tt, err)
		assert.Len(tt, response, 0)
	})
}

func TestRetrieveHandler(t *testing.T) {
	h := newHandler(t)

	t.Run("retrieves cluster successfully", func(tt *testing.T) {
		test.TruncateTables(tt, h.app.DB)
		clusters := []model.Cluster{defaultTestCluster}
		err := h.app.DB.Insert(&clusters)
		require.NoError(tt, err)

		c, rr := test.NewContext(tt, "GET", "", strings.NewReader(""), "application/json")
		c.SetParamNames("id")
		c.SetParamValues(defaultTestCluster.ID)

		err = h.retrieve(c)
		assert.NoError(tt, err)
		assert.Equal(tt, http.StatusOK, rr.Code)

		var response model.Cluster
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(tt, err)
		assert.Equal(tt, defaultTestCluster.ID, response.ID)
		assert.Equal(tt, false, response.Deleted)
	})

	t.Run("retrieves deleted cluster successfully", func(tt *testing.T) {
		test.TruncateTables(tt, h.app.DB)
		clusters := []model.Cluster{deletedTestCluster}
		err := h.app.DB.Insert(&clusters)
		require.NoError(tt, err)

		c, rr := test.NewContext(tt, "GET", "", strings.NewReader(""), "application/json")
		c.SetParamNames("id")
		c.SetParamValues(deletedTestCluster.ID)

		err = h.retrieve(c)
		assert.NoError(tt, err)
		assert.Equal(tt, http.StatusOK, rr.Code)

		var response model.Cluster
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(tt, err)
		assert.Equal(tt, deletedTestCluster.ID, response.ID)
		assert.Equal(tt, true, response.Deleted)
	})

	t.Run("errors retrieves non-existing cluster", func(tt *testing.T) {
		test.TruncateTables(tt, h.app.DB)

		c, _ := test.NewContext(tt, "GET", "", strings.NewReader(""), "application/json")
		c.SetParamNames("id")
		c.SetParamValues("random")

		err := h.retrieve(c)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "not found")
	})
}

func TestDeleteHandler(t *testing.T) {
	h := newHandler(t)

	t.Run("successfully deletes cluster", func(tt *testing.T) {
		test.TruncateTables(tt, h.app.DB)
		clusters := []model.Cluster{defaultTestCluster}
		err := h.app.DB.Insert(&clusters)
		require.NoError(tt, err)

		c, rr := test.NewContext(tt, "DELETE", "", strings.NewReader(""), "application/json")
		c.SetParamNames("id")
		c.SetParamValues(defaultTestCluster.ID)

		err = h.delete(c)
		assert.NoError(tt, err)

		var response model.Cluster
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(tt, err)
		assert.Equal(tt, defaultTestCluster.ID, response.ID)
		assert.Equal(tt, true, response.Deleted)

		var cluster model.Cluster
		err = h.app.DB.Model(&cluster).Where("id = ?", defaultTestCluster.ID).First()
		require.NoError(tt, err)
		assert.True(tt, cluster.Deleted)
	})

	t.Run("errors deleting non-existing cluster", func(tt *testing.T) {
		test.TruncateTables(tt, h.app.DB)

		c, _ := test.NewContext(tt, "DELETE", "", strings.NewReader(""), "application/json")
		c.SetParamNames("id")
		c.SetParamValues("random")

		err := h.delete(c)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "not found")
	})
}

func TestCreateHandler(t *testing.T) {
	h := newHandler(t)

	t.Run("successfully creates cluster", func(tt *testing.T) {
		test.TruncateTables(tt, h.app.DB)

		payload := `{"id": "test-create", "environment": "test", "server_url": "http://localhost:6443", "cluster_authority_data": "dGVzdA=="}`

		c, rr := test.NewContext(tt, "POST", "", strings.NewReader(payload), "application/json")

		err := h.create(c)
		assert.NoError(tt, err)
		assert.Equal(tt, http.StatusOK, rr.Code)

		var response model.Cluster
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(tt, err)
		assert.Equal(tt, "test-create", response.ID)
		assert.Equal(tt, "test", response.Environment)
		assert.Equal(tt, "http://localhost:6443", response.ServerURL)
		assert.Equal(tt, "dGVzdA==", response.ClusterAuthorityData)
		assert.Equal(tt, false, response.Deleted)
		assert.Equal(tt, false, response.Active)
	})

	t.Run("errors with invalid payload", func(tt *testing.T) {
		test.TruncateTables(tt, h.app.DB)

		cases := []struct {
			payload, errorMessage string
		}{
			{
				`{"environment": "test", "server_url": "http://localhost:6443", "cluster_authority_data": "dGVzdA=="}`,
				"id is required",
			},
			{
				`{"id": "test-create", "server_url": "http://localhost:6443", "cluster_authority_data": "dGVzdA=="}`,
				"environment is required",
			},
			{
				`{"id": "test-create", "environment": "test", "cluster_authority_data": "dGVzdA=="}`,
				"server_url is required",
			},
			{
				`{"id": "test-create", "environment": "test", "server_url": "http://localhost:6443"}`,
				"cluster_authority_data is required",
			},
			{
				`{"id": "test-create-invalid-url", "environment": "test", "server_url": "string", "cluster_authority_data": "dGVzdA=="}`,
				"server_url must be a valid URL",
			},
			{
				`{"id": "test-create-invalid-data", "environment": "test", "server_url": "http://localhost:6443", "cluster_authority_data": "!@#$"}`,
				"cluster_authority_data must be a valid base64 encoded string",
			},
		}

		for _, tc := range cases {
			c, _ := test.NewContext(tt, "POST", "", strings.NewReader(tc.payload), "application/json")
			err := h.create(c)
			assert.Error(tt, err)
			fmt.Println(err)
			assert.Contains(tt, err.Error(), tc.errorMessage)
		}
	})

}

func TestUpdateHandler(t *testing.T) {
	h := newHandler(t)

	t.Run("updates clusters successfully", func(tt *testing.T) {
		test.TruncateTables(tt, h.app.DB)
		clusters := []model.Cluster{defaultTestCluster}
		err := h.app.DB.Insert(&clusters)
		require.NoError(tt, err)

		c, rr := test.NewContext(tt, "POST", "", strings.NewReader(`{"active": true}`), "application/json")
		c.SetParamNames("id")
		c.SetParamValues(defaultTestCluster.ID)

		err = h.update(c)
		assert.NoError(tt, err)
		assert.Equal(tt, http.StatusOK, rr.Code)

		var response model.Cluster
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(tt, err)
		assert.True(tt, response.Active)

		var fetchedCluster model.Cluster
		err = h.app.DB.Model(&fetchedCluster).Where("id = ?", defaultTestCluster.ID).First()
		require.NoError(tt, err)
		assert.True(tt, fetchedCluster.Active)
	})

	t.Run("updates clusters successfully and deactivates other clusters", func(tt *testing.T) {
		test.TruncateTables(tt, h.app.DB)
		clusters := []model.Cluster{defaultTestCluster, activeTestCluster}
		err := h.app.DB.Insert(&clusters)
		require.NoError(tt, err)

		c, rr := test.NewContext(tt, "POST", "", strings.NewReader(`{"active": true}`), "application/json")
		c.SetParamNames("id")
		c.SetParamValues(defaultTestCluster.ID)

		err = h.update(c)
		assert.NoError(tt, err)
		assert.Equal(tt, http.StatusOK, rr.Code)

		var response model.Cluster
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(tt, err)
		assert.True(tt, response.Active)

		var fetchedClusters []model.Cluster
		err = h.app.DB.Model(&fetchedClusters).Order("id").Select()
		require.NoError(tt, err)
		assert.Equal(tt, defaultTestCluster.ID, fetchedClusters[0].ID)
		assert.True(tt, fetchedClusters[0].Active)
		assert.Equal(tt, activeTestCluster.ID, fetchedClusters[1].ID)
		assert.False(tt, fetchedClusters[1].Active)
	})

	t.Run("errors updated non-existent cluster", func(tt *testing.T) {
		test.TruncateTables(tt, h.app.DB)

		c, _ := test.NewContext(tt, "POST", "", strings.NewReader(`{"active": true}`), "application/json")
		c.SetParamNames("id")
		c.SetParamValues(defaultTestCluster.ID)

		err := h.update(c)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "cluster not found")
	})
}

func newHandler(t *testing.T) handler {
	t.Helper()

	app, err := application.New()
	require.NoError(t, err)
	return handler{app}
}
