package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/lob/pharos/pkg/pharos/api"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// DeleteCmd implements a CLI command that allows users to mark a cluster
// as deleted in the Pharos database.
var DeleteCmd = &cobra.Command{
	Use:   "delete <cluster_id>",
	Short: "Deletes the specified cluster.",
	Long:  "Marks the specified cluster as deleted in Pharos.",
	Args:  func(cmd *cobra.Command, args []string) error { return argID(args) },
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.ClientFromConfig(pharosConfig)
		if err != nil {
			return errors.Wrap(err, "unable to create client from pharos config file")
		}
		return runDelete(args[0], client)
	},
}

func runDelete(id string, client *api.Client) error {
	cluster, err := client.DeleteCluster(id)
	if err != nil {
		return err
	}
	fmt.Printf("%s CLUSTER %s DELETED\n", color.GreenString("SUCCESS:"), cluster.ID)
	return nil
}
