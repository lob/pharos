package model

import "time"

// Cluster contains a single cluster object returned from
// the pharos API server.
type Cluster struct {
	ID                   string    `json:"id"`
	Environment          string    `json:"environment"`
	ServerURL            string    `json:"server_url"`
	ClusterAuthorityData string    `json:"cluster_authority_data"`
	Deleted              bool      `json:"deleted" sql:",notnull"`
	Active               bool      `json:"active" sql:",notnull"`
	DateCreated          time.Time `json:"date_created"`
	DateModified         time.Time `json:"date_modified"`
}
