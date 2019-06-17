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
	cluster1 = model.Cluster{
		ID:                   "test-1",
		Environment:          "test",
		ServerURL:            "http://test-1.localhost:6443",
		ClusterAuthorityData: "abcdef",
	}
	cluster2 = model.Cluster{
		ID:                   "test-2",
		Environment:          "test",
		ServerURL:            "http://test-2.localhost:6443",
		ClusterAuthorityData: "abcdef",
	}
	cluster3 = model.Cluster{
		ID:                   "test-3",
		Environment:          "test",
		ServerURL:            "http://test-3.localhost:6443",
		ClusterAuthorityData: "abcdef",
		Deleted:              true,
	}
)

func TestListHandler(t *testing.T) {
	h := newHandler(t)

	t.Run("lists clusters ordered correctly", func(tt *testing.T) {
		test.TruncateTables(tt, h.app.DB)
		clusters := []model.Cluster{cluster1, cluster2, cluster3}
		err := h.app.DB.Insert(&clusters)
		require.NoError(tt, err)

		c, rr := test.NewContext(tt, "GET", strings.NewReader(""), "application/json")

		err = h.list(c)
		assert.NoError(tt, err)
		assert.Equal(tt, http.StatusOK, rr.Code)

		var response []model.Cluster
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(tt, err)
		assert.Len(tt, response, 2)
		assert.Equal(tt, cluster1.ID, response[0].ID)
		assert.Equal(tt, cluster2.ID, response[1].ID)
	})

	t.Run("does not list deleted clusters", func(tt *testing.T) {
		test.TruncateTables(tt, h.app.DB)
		clusters := []model.Cluster{cluster3}
		err := h.app.DB.Insert(&clusters)
		require.NoError(tt, err)

		c, rr := test.NewContext(tt, "GET", strings.NewReader(""), "application/json")

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
		clusters := []model.Cluster{cluster1}
		err := h.app.DB.Insert(&clusters)
		require.NoError(tt, err)

		c, rr := test.NewContext(tt, "GET", strings.NewReader(""), "application/json")
		c.SetParamNames("id")
		c.SetParamValues(cluster1.ID)

		err = h.retrieve(c)
		assert.NoError(tt, err)
		assert.Equal(tt, http.StatusOK, rr.Code)

		var response model.Cluster
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(tt, err)
		assert.Equal(tt, cluster1.ID, response.ID)
		assert.Equal(tt, false, response.Deleted)
	})

	t.Run("retrieves deleted cluster successfully", func(tt *testing.T) {
		test.TruncateTables(tt, h.app.DB)
		clusters := []model.Cluster{cluster3}
		err := h.app.DB.Insert(&clusters)
		require.NoError(tt, err)

		c, rr := test.NewContext(tt, "GET", strings.NewReader(""), "application/json")
		c.SetParamNames("id")
		c.SetParamValues(cluster3.ID)

		err = h.retrieve(c)
		assert.NoError(tt, err)
		assert.Equal(tt, http.StatusOK, rr.Code)

		var response model.Cluster
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(tt, err)
		assert.Equal(tt, cluster3.ID, response.ID)
		assert.Equal(tt, true, response.Deleted)
	})

	t.Run("errors retrieves non-existing cluster", func(tt *testing.T) {
		test.TruncateTables(tt, h.app.DB)

		c, _ := test.NewContext(tt, "GET", strings.NewReader(""), "application/json")
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
		clusters := []model.Cluster{cluster1}
		err := h.app.DB.Insert(&clusters)
		require.NoError(tt, err)

		c, rr := test.NewContext(tt, "DELETE", strings.NewReader(""), "application/json")
		c.SetParamNames("id")
		c.SetParamValues(cluster1.ID)

		err = h.delete(c)
		assert.NoError(tt, err)

		var response model.Cluster
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(tt, err)
		assert.Equal(tt, cluster1.ID, response.ID)
		assert.Equal(tt, true, response.Deleted)

		var cluster model.Cluster
		err = h.app.DB.Model(&cluster).Where("id = ?", cluster1.ID).First()
		require.NoError(tt, err)
		assert.True(tt, cluster.Deleted)
	})

	t.Run("errors deleting non-existing cluster", func(tt *testing.T) {
		test.TruncateTables(tt, h.app.DB)

		c, _ := test.NewContext(tt, "DELETE", strings.NewReader(""), "application/json")
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

		c, rr := test.NewContext(tt, "POST", strings.NewReader(payload), "application/json")

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
			c, _ := test.NewContext(tt, "POST", strings.NewReader(tc.payload), "application/json")
			err := h.create(c)
			assert.Error(tt, err)
			fmt.Println(err)
			assert.Contains(tt, err.Error(), tc.errorMessage)
		}
	})

}

func newHandler(t *testing.T) handler {
	t.Helper()

	app, err := application.New()
	require.NoError(t, err)
	return handler{app}
}
