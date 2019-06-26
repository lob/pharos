package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/lob/pharos/pkg/pharos/api"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Declare some variables to be used as flags.
var active bool

// UpdateCmd implements a CLI command that allows users to update the status of
// a cluster in Pharos.
var UpdateCmd = &cobra.Command{
	Use:   "update <cluster_id>",
	Short: "Updates the status of a cluster",
	Long:  "Updates the status of the specified cluster in Pharos.",
	Args:  func(cmd *cobra.Command, args []string) error { return argID(args) },
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.ClientFromConfig(pharosConfig)
		if err != nil {
			return errors.Wrap(err, "unable to create client from pharos config file")
		}
		return runUpdate(args[0], active, client)
	},
}

func runUpdate(id string, active bool, client *api.Client) error {
	cluster, err := client.UpdateCluster(id, active)
	if err != nil {
		return err
	}
	fmt.Printf("%s CLUSTER %s ACTIVE STATUS UPDATED TO %t\n", color.GreenString("SUCCESS:"), cluster.ID, cluster.Active)
	return nil
}

func init() {
	UpdateCmd.Flags().BoolVarP(&active, "active", "a", true, "specify whether to set the cluster status to active")
}
