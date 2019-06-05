package cmd

import (
	"fmt"

	"github.com/lob/pharos/pkg/pharos/kubeconfig"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var file string

// ClustersCmd is the pharos clusters command.
var ClustersCmd = &cobra.Command{
	Use:   "clusters",
	Short: "Print current cluster",
	Long: `Prints current cluster from the context in the kubeconfig file at
$HOME/.kube/config, unless otherwise specified.`,
	RunE: func(cmd *cobra.Command, args []string) error { return runClusters(file) },
}

func runClusters(kubeConfigFile string) error {
	clusterName, err := kubeconfig.CurrentCluster(kubeConfigFile)
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve cluster")
	}
	fmt.Println(clusterName)
	return nil
}

func init() {
	ClustersCmd.Flags().StringVarP(&file, "file", "f", "", "specify kubeconfig file")
}
