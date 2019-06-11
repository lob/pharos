package api

import (
	"fmt"
	"net/http"

	"github.com/lob/pharos/pkg/util/model"
	"github.com/pkg/errors"
)

// ListClusters sends a GET request to the clusters endpoint of the Pharos API
// and returns an array of Clusters.
func (c *Client) ListClusters() ([]model.Cluster, error) {
	var clusters []model.Cluster
	err := c.send(http.MethodGet, "clusters", nil, &clusters)
	if err != nil {
		return clusters, errors.Wrap(err, "failed to list clusters")
	}

	return clusters, nil
}

// GetCluster sends a GET request to the clusters/id endpoint of the Pharos API
// and returns a Cluster.
func (c *Client) GetCluster(clusterID string) (model.Cluster, error) {
	var cluster model.Cluster
	err := c.send(http.MethodGet, fmt.Sprintf("clusters/%s", clusterID), nil, &cluster)
	if err != nil {
		return cluster, errors.Wrap(err, "failed to get cluster")
	}

	return cluster, nil
}
