package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Declare some variables to be used as flags in various commands.
var (
	dryRun        bool
	environment   string
	file          string
	inactive      bool
	pharosConfig  string
	pharosVersion string // pharosVersion can be overwritten by ldflags in the Makefile.
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:     "pharos",
	Short:   "A tool for managing kubeconfig files.",
	Long:    "Pharos is a tool for cluster discovery and distribution of kubeconfig files.",
	Version: pharosVersion,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Set default version number if none given.
	if rootCmd.Version == "" {
		rootCmd.Version = "0.0.0"
	}

	rootCmd.Flags().BoolP("version", "v", false, "print Pharos version number")
	rootCmd.PersistentFlags().StringVarP(&pharosConfig, "config", "c", fmt.Sprintf("%s/.kube/pharos/config", os.Getenv("HOME")), "Pharos config file")

	// Prevent usage message from being printed out upon command error.
	rootCmd.SilenceUsage = true

	// Add child commands.
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(NewClustersCmd())
	rootCmd.AddCommand(SetupCmd)
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
