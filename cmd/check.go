package cmd

import (
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check various aspects of your AWS resources",
	Long: `Check command provides various subcommands to verify and check the status
of your AWS resources, including:
- credentials: Verify AWS credentials and assume roles
- migrate: Check if your instances need migration to newer AMIs`,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
