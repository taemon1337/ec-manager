package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/client"
)

// Common flags and variables
var (
	// Mock mode flag
	mockMode bool

	// AWS client
	awsClient *client.Client
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ecman",
	Short: "EC2 Manager CLI",
	Long:  `A CLI tool for managing EC2 instances`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg := client.NewDefaultConfig()
		cfg.MockMode = mockMode
		awsClient, err = client.NewClient(cfg)
		if err != nil {
			return fmt.Errorf("failed to create AWS client: %w", err)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		// Find the command
		c, _, err := cmd.Root().Find(args)
		if err != nil {
			return fmt.Errorf("unknown command %q", args[0])
		}
		return c.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&mockMode, "mock", false, "Enable mock mode")
}
