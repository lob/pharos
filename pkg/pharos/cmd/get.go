package cmd

import (
	"fmt"
	"os"

	"github.com/lob/pharos/pkg/pharos/api"
	"github.com/lob/pharos/pkg/pharos/kubeconfig"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Declare some variables to be used as flags.
var dryRun bool

// GetCmd implements a CLI command that allows users to get cluster information from a new cluster
// and merge it into an existing kubeconfig file.
var GetCmd = &cobra.Command{
	Use:   "get <cluster_id>",
	Short: "Retrieves information about the specified cluster",
	Long:  "Retrieves information about the specified cluster and merges it into designated kubeconfig file.",
	Args:  func(cmd *cobra.Command, args []string) error { return argID(args) },
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.ClientFromConfig(pharosConfig)
		if err != nil {
			return errors.Wrap(err, "unable to create client from pharos config file")
		}
		return runGet(args[0], file, dryRun, client)
	},
}

func runGet(cluster string, kubeConfigFile string, dryRun bool, client *api.Client) error {
	err := kubeconfig.GetCluster(cluster, kubeConfigFile, dryRun, client)
	if err != nil {
		return errors.Wrap(err, "failed to get cluster information")
	}
	return nil
}

func init() {
	GetCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "prints the resulting kubeconfig to terminal without any other action")
	GetCmd.Flags().StringVarP(&file, "file", "f", fmt.Sprintf("%s/.kube/config", os.Getenv("HOME")), "specify kubeconfig file to merge into")
}
