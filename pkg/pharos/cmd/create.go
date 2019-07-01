package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/lob/pharos/pkg/pharos/api"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Declare some variables to be used as flags.
var (
	clusterAuthorityData string
	server               string
)

// CreateCmd implements a CLI command that allows users to create a cluster in Pharos.
var CreateCmd = &cobra.Command{
	Use:     "create <cluster_id>",
	Short:   "Creates the specified cluster",
	Long:    "Creates the specified cluster in Pharos.",
	Args:    func(cmd *cobra.Command, args []string) error { return argID(args) },
	PreRunE: func(cmd *cobra.Command, args []string) error { return markFlagsRequired(cmd) },
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.ClientFromConfig(pharosConfig)
		if err != nil {
			return errors.Wrap(err, "unable to create client from pharos config file")
		}
		return runCreate(args[0], environment, clusterAuthorityData, server, client)
	},
}

func runCreate(id string, env string, authorityData string, server string, client *api.Client) error {
	newCluster := api.NewCluster{
		ID:                   id,
		Environment:          env,
		ClusterAuthorityData: authorityData,
		ServerURL:            server,
	}
	cluster, err := client.CreateCluster(newCluster)
	if err != nil {
		return err
	}
	fmt.Printf("%s CREATED CLUSTER %s\n", color.GreenString("SUCCESS:"), cluster.ID)
	return nil
}

func markFlagsRequired(cmd *cobra.Command) error {
	createFlags := cmd.Flags()
	if err := cobra.MarkFlagRequired(createFlags, "environment"); err != nil {
		return err
	}
	if err := cobra.MarkFlagRequired(createFlags, "cluster-authority-data"); err != nil {
		return err
	}
	if err := cobra.MarkFlagRequired(createFlags, "server"); err != nil {
		return err
	}
	return nil
}

func init() {
	CreateCmd.Flags().StringVarP(&environment, "environment", "e", "", "environment of the cluster (required)")
	CreateCmd.Flags().StringVarP(&clusterAuthorityData, "cluster-authority-data", "d", "", "cluster authority data of the cluster (required)")
	CreateCmd.Flags().StringVarP(&server, "server", "s", "", "server url of the cluster (required)")
}
