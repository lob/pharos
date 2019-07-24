package cmd

import (
	"github.com/spf13/cobra"
)

// NewClustersCmd returns a new cobra.Command with all the necessary clusters
// sub-commands attached to it.
func NewClustersCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "clusters",
		Short: `Commands for cluster management (run "pharos clusters -h" for a full list of cluster commands)`,
		Long:  "Commands for managing cluster kubeconfig files.",
	}

	cmd.AddCommand(CreateCmd)
	cmd.AddCommand(CurrentCmd)
	cmd.AddCommand(DeleteCmd)
	cmd.AddCommand(GetCmd)
	cmd.AddCommand(ListCmd)
	cmd.AddCommand(SwitchCmd)
	cmd.AddCommand(SyncCmd)
	cmd.AddCommand(UpdateCmd)

	return cmd
}
