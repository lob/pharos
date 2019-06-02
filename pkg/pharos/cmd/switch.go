package cmd

import (
	"fmt"
	"os"

	"github.com/lob/pharos/pkg/pharos/kubeconfig"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// SwitchCmd implements a CLI command that allows users to switch between clusters.
var SwitchCmd = &cobra.Command{
	Use:   "switch <cluster_id>",
	Short: "Switch to specified cluster",
	Long:  "Switches the current context in the designated kubeconfig file to the context referencing the specified cluster.",
	Args:  func(cmd *cobra.Command, args []string) error { return argID(args) },
	RunE:  func(cmd *cobra.Command, args []string) error { return runSwitch(file, args[0]) },
}

func runSwitch(kubeConfigFile string, context string) error {
	fmt.Printf("SWITCHING TO CLUSTER %s...\n", context)

	err := kubeconfig.SwitchCluster(kubeConfigFile, context)
	if err != nil {
		return errors.Wrap(err, "cluster switch unsuccessful")
	}

	fmt.Println("CLUSTER SWITCH COMPLETE.")
	return nil
}

func init() {
	SwitchCmd.Flags().StringVarP(&file, "file", "f", fmt.Sprintf("%s%s", os.Getenv("HOME"), "/.kube/config"), "specify designated kubeconfig file")
}
