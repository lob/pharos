package kubeconfig

import (
	"fmt"
	"os"
	"strings"

	"github.com/lob/pharos/pkg/pharos/api"
	"github.com/lob/pharos/pkg/util/model"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// CurrentCluster returns current context name.
func CurrentCluster(kubeConfigFile string) (string, error) {
	kubeConfig, err := configFromFile(kubeConfigFile)
	if err != nil {
		return "", errors.Wrap(err, "unable to load kubeconfig file")
	}

	// Make sure that current context exists.
	_, ok := kubeConfig.Contexts[kubeConfig.CurrentContext]
	if !ok {
		return "", errors.New("context not found")
	}

	return kubeConfig.CurrentContext, nil
}

// GetCluster gets information from a new cluster
// and merges it into an existing kubeconfig file
func GetCluster(id string, kubeConfigFile string, dryRun bool, client *api.Client) error {
	// Check whether given kubeconfig file already exists. If it does not, create a new kubeconfig
	// file in the specified file location. Return an error only if file is malformed, but not
	// if it is empty or missing.
	kubeConfig, err := configFromFile(kubeConfigFile)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "unable to load kubeconfig file")
	}
	if kubeConfig == nil {
		kubeConfig = clientcmdapi.NewConfig()
	}

	// Check whether given id has a suffix. If not, this means that we were
	// given an environment name instead of an actual cluster id and we need to
	// fetch the cluster id of the currently active cluster.
	var cluster model.Cluster
	if !strings.Contains(id, "-") {
		// Create query to find active cluster of given environment.
		q := map[string]string{
			"active":      "true",
			"environment": id,
		}

		clusters, err := client.ListClusters(q)
		if err != nil {
			return errors.Wrap(err, "unable to list clusters for specified environment")
		}
		switch {
		case len(clusters) < 1:
			return fmt.Errorf("no active cluster found for environment %s", id)
		case len(clusters) > 1:
			return fmt.Errorf("multiple clusters found for environment %s", id)
		}

		cluster = clusters[0]
	} else {
		// Get cluster information for a specific cluster from Pharos API.
		cluster, err = client.GetCluster(id)
		if err != nil {
			return err
		}
	}

	// If a kubeconfig has no current context set, set current context to the environment or
	// id that was passed in.
	if kubeConfig.CurrentContext == "" {
		kubeConfig.CurrentContext = id
	}

	// Set username associated with new cluster.
	clusterID := cluster.ID
	username := fmt.Sprintf("engineering-%s", clusterID)

	// Update user, context, and cluster information associated with the cluster
	// in the kubeconfig.
	kubeConfig.Clusters[clusterID] = newCluster(cluster)
	kubeConfig.AuthInfos[username] = newUser(clusterID)
	context := newContext(clusterID, username)
	kubeConfig.Contexts[clusterID] = context

	// If necessary, update existing context for the specified environment.
	kubeConfig.Contexts[id] = context

	// Check for errors in newly created config.
	err = clientcmd.Validate(*kubeConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create valid kubeconfig")
	}

	// Print kubeconfig to terminal instead of saving to file during a dry run.
	if dryRun {
		yaml, err := clientcmd.Write(*kubeConfig)
		if err != nil {
			return errors.Wrap(err, "unable to write kubeconfig file")
		}
		fmt.Println(string(yaml))
		return nil
	}

	return clientcmd.WriteToFile(*kubeConfig, kubeConfigFile)
}

// SwitchCluster switches current context to given cluster or context name.
func SwitchCluster(kubeConfigFile string, context string) error {
	kubeConfig, err := configFromFile(kubeConfigFile)
	if err != nil {
		return err
	}

	// Check if there is a context corresponding to the given context name or cluster.
	_, ok := kubeConfig.Contexts[context]
	if !ok {
		return errors.New("cluster does not exist in context")
	}

	// Switch to new cluster.
	kubeConfig.CurrentContext = context

	// Check for errors in new config.
	err = clientcmd.Validate(*kubeConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create valid kubeconfig")
	}

	return clientcmd.WriteToFile(*kubeConfig, kubeConfigFile)
}

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
func newUser(id string) *clientcmdapi.AuthInfo {
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
	env.Value = strings.Split(id, "-")[0]
	exec.Env = []clientcmdapi.ExecEnvVar{env}

	return user
}

// newCluster returns a pointer to a new clientcmdapi.Cluster containing
// information from a cluster.
func newCluster(c model.Cluster) *clientcmdapi.Cluster {
	cluster := clientcmdapi.NewCluster()
	cluster.Server = c.ServerURL
	cluster.CertificateAuthorityData = []byte(c.ClusterAuthorityData)
	return cluster
}
