package cli

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"text/tabwriter"

	"github.com/fatih/color"
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
// and merges it into an existing kubeconfig file.
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

	// Check whether given id has a suffix composed of a dash followed by six numbers.
	// (Example: sandbox-111111 vs sandbox)
	// If the id has no suffix, this means that we were given an environment name instead
	// of a cluster id and we need to fetch the id of the currently active cluster from
	// the Pharos API.
	var cluster model.Cluster
	match, err := regexp.MatchString(`-\d{6}`, id)
	if err != nil {
		return errors.Wrap(err, "unable to match cluster ID with regex")
	}
	if !match {
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
			return fmt.Errorf("%d clusters found for environment %s", len(clusters), id)
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
	username := fmt.Sprintf("iam-%s", clusterID)

	// Update user, context, and cluster information associated with the cluster
	// in the kubeconfig.
	kubeConfig.Clusters[clusterID] = newCluster(cluster)
	kubeConfig.AuthInfos[username] = newUser(clusterID, cluster.Environment)
	context := newContext(clusterID, username)
	kubeConfig.Contexts[clusterID] = context

	// Update existing context for the specified environment.
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

	// Write kubeconfig to file.
	err = clientcmd.WriteToFile(*kubeConfig, kubeConfigFile)
	if err != nil {
		return err
	}
	fmt.Printf("%s %s MERGED INTO %s\n", color.GreenString("SUCCESS:"), id, kubeConfigFile)
	return nil
}

// ListClusters retrieves all clusters and returns a formatted string
// of all clusters. If given an environment, ListClusters will only retrieve
// the clusters for that environment.
func ListClusters(env string, inactive bool, client *api.Client) (string, error) {
	query := make(map[string]string)
	// If inactive is false, we'll only list active clusters, otherwise we'll list
	// all clusters, including inactive ones.
	if !inactive {
		query["active"] = "true"
	}
	if env != "" {
		query["environment"] = env
	}

	c, err := client.ListClusters(query)
	if err != nil {
		return "", err
	}

	// List cluster attributes in organized columns.
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 0, 3, ' ', 0)
	cyan := color.New(color.FgCyan)

	// Add spaces to prevent ANSI escape codes from breaking the tabwriter formatting.
	_, err = cyan.Fprint(w, "CLUSTER_ID\t     ENVIRONMENT\t     ACTIVE\t     SERVER")
	if err != nil {
		return "", err
	}

	for _, cluster := range c {
		fmt.Fprintf(w, "\n%s\t%s\t%s\t%s", cluster.ID, cluster.Environment, strconv.FormatBool(cluster.Active), cluster.ServerURL)
	}

	fmt.Fprintln(w, "")
	if err := w.Flush(); err != nil {
		return "", err
	}

	return buf.String(), nil
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

// SyncClusters gets information from all current existing clusters
// and merges it into a kubeconfig file.
func SyncClusters(kubeConfigFile string, dryRun bool, overwrite bool, client *api.Client) error {
	// Check whether given kubeconfig file already exists. If it does not, create a new kubeconfig
	// file in the specified file location. Return an error only if file is malformed, but not
	// if it is empty or missing. If overwrite is set to true, start with new kubeconfig file
	// regardless.
	kubeConfig, err := configFromFile(kubeConfigFile)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "unable to load kubeconfig file")
	}
	if kubeConfig == nil || overwrite {
		kubeConfig = clientcmdapi.NewConfig()
	}

	clusters, err := client.ListClusters(nil)
	if err != nil {
		return err
	}

	// Add cluster, context, and user for each cluster. There should never be
	// more than one cluster marked active for each environment, but if there is,
	// return an error.
	active := map[string]bool{}
	for _, cluster := range clusters {
		clusterID := cluster.ID
		env := cluster.Environment
		username := fmt.Sprintf("iam-%s", clusterID)
		kubeConfig.Clusters[clusterID] = newCluster(cluster)
		kubeConfig.AuthInfos[username] = newUser(clusterID, env)
		context := newContext(clusterID, username)
		kubeConfig.Contexts[clusterID] = context

		if cluster.Active {
			_, ok := active[env]
			if ok {
				return fmt.Errorf("more than one active cluster for environment %s found", env)
			}
			kubeConfig.Contexts[env] = context
			active[env] = true
		}
	}

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

	err = clientcmd.WriteToFile(*kubeConfig, kubeConfigFile)
	if err != nil {
		return err
	}
	fmt.Printf("%s %d CLUSTERS SYNCED AND MERGED INTO %s\n", color.GreenString("SUCCESS:"), len(clusters), kubeConfigFile)
	return nil
}
