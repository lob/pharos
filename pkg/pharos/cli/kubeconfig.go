package cli

import (
	"encoding/base64"

	"github.com/lob/pharos/pkg/util/model"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// configFromFile returns a struct containing kubeconfig information from a file.
// Does not differentiate between errors resulting from a missing file and errors
// from reading from a malformed config.
// Function source: https://github.com/kubernetes/client-go/blob/88ff0afc48bbf242f66f2f0c8d5c26b253e6561c/tools/clientcmd/config.go#L471
func configFromFile(fileName string) (*clientcmdapi.Config, error) {
	kubeConfig, err := clientcmd.LoadFromFile(fileName)
	if err != nil {
		return nil, err
	}
	return kubeConfig, nil
}

// newContext returns a pointer to a new kubeconfig context with specified cluster and user.
func newContext(id string, user string) *clientcmdapi.Context {
	context := clientcmdapi.NewContext()
	context.Cluster = id
	context.AuthInfo = user

	return context
}

// newUser returns a pointer to a new kubeconfig user for a specified cluster.
// The id given should always be of form "[environment]-[suffix]".
func newUser(id string, environment string) *clientcmdapi.AuthInfo {
	user := clientcmdapi.NewAuthInfo()

	// Add exec config.
	var exec clientcmdapi.ExecConfig
	exec.Command = "aws-iam-authenticator"
	exec.APIVersion = "client.authentication.k8s.io/v1alpha1"
	exec.Args = []string{"token", "-i", id}
	user.Exec = &exec

	// Add env variables to exec config.
	var env clientcmdapi.ExecEnvVar
	env.Name = "AWS_PROFILE"
	env.Value = environment
	exec.Env = []clientcmdapi.ExecEnvVar{env}

	return user
}

// newCluster returns a pointer to a new clientcmdapi.Cluster containing
// information from a cluster.
func newCluster(c model.Cluster) *clientcmdapi.Cluster {
	clusterAuthorityData, _ := base64.StdEncoding.DecodeString(c.ClusterAuthorityData)

	cluster := clientcmdapi.NewCluster()
	cluster.Server = c.ServerURL
	cluster.CertificateAuthorityData = clusterAuthorityData
	return cluster
}
