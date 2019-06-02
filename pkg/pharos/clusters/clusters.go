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
		clusterName, err := kubeconfig.CurrentCluster(file)
		if err != nil {
			fmt.Printf("ERROR: Unable to retrieve cluster - %s\n", err.Error())
		} else {
			fmt.Println(clusterName)
		}
	},
}

func init() {
	ClustersCmd.Flags().StringVarP(&file, "file", "f", "", "specify kubeconfig file")
}
