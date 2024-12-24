package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Common flags and variables
var (
	mockMode   bool
	instanceID string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ecman <action>",
	Short: "EC2 instance management tool",
	Long: `ec-manager (ecman) is a CLI tool that helps you manage your AWS EC2 instances.
It provides commands for:
- Creating new instances with proper configuration
- Listing and checking instance status
- Migrating instances to new AMIs
- Managing instance backups
- Cleaning up unused instances`,
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

func init() {
	rootCmd.PersistentFlags().BoolVar(&mockMode, "mock", false, "Enable mock mode")
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
