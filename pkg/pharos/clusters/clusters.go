package clusters

import (
	"fmt"

	"github.com/lob/pharos/pkg/pharos/kubeconfig"
	"github.com/spf13/cobra"
)

var file string

// ClustersCmd is the pharos clusters command
var ClustersCmd = &cobra.Command{
	Use:   "clusters",
	Short: "Print current cluster",
	Long: `Prints current cluster from the context in the kubeconfig file at
$HOME/.kube/config, unless otherwise specified.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(kubeconfig.CurrentCluster(&file))
	},
}

func init() {
	ClustersCmd.AddCommand(SwitchCmd)
	ClustersCmd.Flags().StringVarP(&file, "file", "f", "/.kube/config", "specify kubeconfig file")
}
