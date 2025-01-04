package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/client"
)

var (
	awsClient *client.Client
	mockMode  bool
	region    string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ec-manager",
	Short: "A CLI tool for managing EC2 instances",
	Long: `A CLI tool for managing EC2 instances, including:
- Creating and managing backups
- Migrating instances to new AMIs
- Restoring instances from backups`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		awsClient, err = client.NewClient(mockMode, "", region)
		if err != nil {
			return fmt.Errorf("failed to create AWS client: %w", err)
		}
		return nil
	},
}

// NewRootCmd creates a new root command
func NewRootCmd() *cobra.Command {
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
		return err
	}
	return nil
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&mockMode, "mock", false, "Use mock mode for testing")
	rootCmd.PersistentFlags().StringVar(&region, "region", "us-east-1", "AWS region to use")
}
