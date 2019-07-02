package cmd

import (
	"fmt"

	"github.com/fatih/color"
	configpkg "github.com/lob/pharos/pkg/pharos/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Declare variables to be used as flags.
var awsProfile string
var awsRoleARN string
var pharosURL string

// setupCmd is the pharos setup command.
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup Pharos config",
	Long:  "Setup Pharos configuration file. Overwrites previously saved configuration.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetup(pharosConfig, pharosURL, awsProfile, awsRoleARN)
	},
}

func runSetup(pharosConfig string, url string, profile string, arn string) error {
	c, err := configpkg.New(pharosConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create reference to config file")
	}

	c.BaseURL = url
	c.AWSProfile = profile
	c.AssumeRoleARN = arn

	err = c.Save()
	if err != nil {
		return errors.Wrap(err, "unable to save config file")
	}
	fmt.Printf("%s SAVED CONFIG TO %s\n", color.GreenString("SUCCESS:"), pharosConfig)
	return nil
}

func init() {
	setupCmd.Flags().StringVarP(&awsProfile, "aws-profile", "p", "", "specify aws profile to use")
	setupCmd.Flags().StringVarP(&awsRoleARN, "aws-role-arn", "r", "", "specify aws role ARN to use")
	setupCmd.Flags().StringVarP(&pharosURL, "pharos-url", "u", "", "specify URL of the pharos server")
}
