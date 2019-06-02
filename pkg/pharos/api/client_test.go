package api

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lob/pharos/pkg/pharos/config"
	"github.com/lob/pharos/pkg/util/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const configFile = "../testdata/pharosConfig"

func TestClient(t *testing.T) {
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

	t.Run("successfully creates a new client", func(tt *testing.T) {
		c := NewClient(&config.Config{})
		assert.NotNil(tt, c)

		assert.Equal(tt, 10*time.Second, c.client.Timeout)
	})

	t.Run("send makes a successful GET request", func(tt *testing.T) {
		c := NewClient(&config.Config{BaseURL: srv.URL})
		cluster := model.Cluster{}
		// TODO: Test making a GET request to the pharos-api-server when that's been set up.
		// c := NewClient(Config{BaseURL: "http://localhost:7654"})

		err := c.send(http.MethodGet, "", nil, &cluster, nil)
		assert.NoError(tt, err)
		assert.Equal(tt, "production-6906ce", cluster.ID)
	})

	t.Run("correctly bubbles up HTTP errors", func(tt *testing.T) {
		c := NewClient(&config.Config{BaseURL: "bad url"})
		cluster := model.Cluster{}

		err := c.send("", "", nil, &cluster, nil)
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "unsupported protocol")
	})
}

func TestClientFromConfig(t *testing.T) {
	t.Run("successfully creates a new client", func(tt *testing.T) {
		c, err := ClientFromConfig(configFile)
		require.NoError(tt, err)
		assert.NotNil(tt, c)

		assert.Equal(tt, 10*time.Second, c.client.Timeout)
		assert.Equal(tt, "pharos.lob-sandbox.com", c.config.BaseURL)
	})
}

func TestCheckError(t *testing.T) {
	t.Run("fails upon receiving a response with a bad status code", func(tt *testing.T) {
		err := checkError(&http.Response{
			Body:       ioutil.NopCloser(strings.NewReader(`{"error": {"message" : "internal server error"}}`)),
			StatusCode: http.StatusInternalServerError,
		})
		assert.Error(tt, err)
		assert.Contains(tt, err.Error(), "internal server error")
	})

	t.Run("returns nil upon receiving a response with no errors", func(tt *testing.T) {
		err := checkError(&http.Response{
			Body:       ioutil.NopCloser(strings.NewReader("{ok}")),
			StatusCode: http.StatusOK,
		})
		assert.NoError(tt, err)
	})
}
