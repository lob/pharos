package cmd

import (
	"fmt"
	"os"

	"github.com/lob/pharos/pkg/pharos/clusters"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "pharos",
	Short:   "A tool for managing kubeconfig files.",
	Long:    `Pharos is a tool for cluster discovery and distribution of kubeconfig files.`,
	Version: "1.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.Flags().BoolP("version", "v", false, "print pharos version number")

	// Add child commands.
	RootCmd.AddCommand(clusters.ClustersCmd)
}
