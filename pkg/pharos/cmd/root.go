package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Declare some variables to be used as flags in various commands.
var file string
var pharosConfig string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:     "pharos",
	Short:   "A tool for managing kubeconfig files.",
	Long:    `Pharos is a tool for cluster discovery and distribution of kubeconfig files.`,
	Version: "1.0",
}

// clustersCmd is the pharos clusters command.
var clustersCmd = &cobra.Command{Use: "clusters"}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "print pharos version number")
	rootCmd.PersistentFlags().StringVarP(&pharosConfig, "config", "c", fmt.Sprintf("%s/.kube/pharos/config", os.Getenv("HOME")), "pharos config file")

	// Prevent usage message from being printed out upon command error.
	rootCmd.SilenceUsage = true

	// Add child commands.
	rootCmd.AddCommand(clustersCmd)
	clustersCmd.AddCommand(CurrentCmd)
	clustersCmd.AddCommand(SwitchCmd)
	clustersCmd.AddCommand(GetCmd)
}

// argID prevents commands from being run unless exactly one argument (a cluster name or id)
// has been passed in. This function is used in many child commands.
func argID(args []string) error {
	if len(args) < 1 {
		return errors.New("requires a cluster name or id argument")
	}
	if len(args) > 1 {
		return errors.New("too many arguments given")
	}
	return nil
}
