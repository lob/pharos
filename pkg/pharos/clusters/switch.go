package clusters

import (
	"fmt"

	"github.com/spf13/cobra"
)

// SwitchCmd implements a CLI command that allows users to switch between clusters.
var SwitchCmd = &cobra.Command{
	Use:   "switch [ID]",
	Short: "Switch to specified cluster",
	Long: `Switches the current context in the kubeconfig file at $HOME/.kube/config
to the context referencing the specified cluster.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hello called")
	},
}
