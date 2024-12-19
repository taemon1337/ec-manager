package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Shared flags
	instanceID   string
	enabledValue string
	newAMI       string
	latestTag    string
)

var rootCmd = &cobra.Command{
	Use:   "ami-migrate",
	Short: "AMI Migration Tool",
	Long: `A tool for managing AWS EC2 instance AMI migrations, backups, and restores.
Can be used both from command line and in CI/CD pipelines.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&newAMI, "new-ami", "", "ID of the new AMI to migrate to")
	rootCmd.PersistentFlags().StringVar(&latestTag, "latest-tag", "latest", "Tag value for the current latest AMI")
	rootCmd.PersistentFlags().StringVar(&enabledValue, "enabled-value", "enabled", "Value to match for the ami-migrate tag")
}
