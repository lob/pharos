package cmd

import (
	"fmt"

	"github.com/lob/pharos/pkg/pharos/api"
	"github.com/lob/pharos/pkg/pharos/cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Declare some variables to be used as flags.
var environment string

// ListCmd implements a CLI command that allows users to retrieve a list of all clusters
// currently registered with pharos-api.
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "Retrieves a list of all clusters.",
	Long:  "Retrieves a list of all clusters currently registered with Pharos.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.ClientFromConfig(pharosConfig)
		if err != nil {
			return errors.Wrap(err, "unable to create client from pharos config file")
		}
		return runList(environment, active, client)
	},
}

func runList(env string, active bool, client *api.Client) error {
	clusters, err := cli.ListClusters(environment, active, client)
	if err != nil {
		return errors.Wrap(err, "failed to list clusters")
	}
	fmt.Print(clusters)
	return nil
}

func init() {
	ListCmd.Flags().StringVarP(&environment, "environment", "e", "", "specify environment to list clusters for")
	ListCmd.Flags().BoolVarP(&active, "active", "a", true, "status of clusters to list")
}
