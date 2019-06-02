package cmd

import (
	"fmt"

	"github.com/lob/pharos/pkg/pharos/kubeconfig"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// CurrentCmd is the pharos clusters command.
var CurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Print current cluster",
	Long: `Prints current cluster from the context in the kubeconfig file at
$HOME/.kube/config, unless otherwise specified.`,
	RunE: func(cmd *cobra.Command, args []string) error { return runCurrent(file) },
}

func runCurrent(kubeConfigFile string) error {
	clusterName, err := kubeconfig.CurrentCluster(kubeConfigFile)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve cluster")
	}
	fmt.Println(clusterName)
	return nil
}

func init() {
	CurrentCmd.Flags().StringVarP(&file, "file", "f", "", "specify kubeconfig file")
}
