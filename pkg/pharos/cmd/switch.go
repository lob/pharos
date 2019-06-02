package cmd

import (
	"fmt"

	"github.com/lob/pharos/pkg/pharos/kubeconfig"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// SwitchCmd implements a CLI command that allows users to switch between clusters.
var SwitchCmd = &cobra.Command{
	Use:   "switch [ID]",
	Short: "Switch to specified cluster",
	Long: `Switches the current context in the kubeconfig file at $HOME/.kube/config
to the context referencing the specified cluster.`,
	Args: func(cmd *cobra.Command, args []string) error { return argSwitch(args) },
	RunE: func(cmd *cobra.Command, args []string) error { return runSwitch(file, args[0]) },
}

func runSwitch(kubeConfigFile string, context string) error {
	fmt.Println("SWITCHING TO CLUSTER", context+"...")

	err := kubeconfig.SwitchCluster(kubeConfigFile, context)
	if err != nil {
		return errors.Wrap(err, "cluster switch unsuccessful")
	}

	fmt.Println("CLUSTER SWITCH COMPLETE.")
	return nil
}

func argSwitch(args []string) error {
	if len(args) < 1 {
		return errors.New("requires a cluster name or ID argument")
	}
	if len(args) > 1 {
		return errors.New("too many arguments given")
	}
	return nil
}

func init() {
	SwitchCmd.Flags().StringVarP(&file, "file", "f", "", "specify kubeconfig file")
}
