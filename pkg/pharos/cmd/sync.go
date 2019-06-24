package cmd

import (
	"fmt"
	"os"

	"github.com/lob/pharos/pkg/pharos/api"
	"github.com/lob/pharos/pkg/pharos/cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Declare variable to be used as a flag.
var overwrite bool

// SyncCmd implements a CLI command that allows users to get cluster information
// from all currently existing clusters in Pharos and merge it into an existing kubeconfig file.
var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Retrieves information from all clusters",
	Long:  "Retrieves information from all clusters and merges it into designated kubeconfig file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.ClientFromConfig(pharosConfig)
		if err != nil {
			return errors.Wrap(err, "unable to create client from pharos config file")
		}
		return runSync(file, dryRun, overwrite, client)
	},
}

func runSync(kubeConfigFile string, dryRun bool, overwrite bool, client *api.Client) error {
	err := cli.SyncClusters(kubeConfigFile, dryRun, overwrite, client)
	if err != nil {
		return errors.Wrap(err, "failed to sync clusters")
	}
	return nil
}

func init() {
	SyncCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "prints the resulting kubeconfig to terminal without any other action")
	SyncCmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "overwrite the kubeconfig file with retrieved clusters")
	SyncCmd.Flags().StringVarP(&file, "file", "f", fmt.Sprintf("%s/.kube/config", os.Getenv("HOME")), "specify kubeconfig file to merge into")
}
