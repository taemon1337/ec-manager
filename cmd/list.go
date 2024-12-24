package cmd

import (
	"github.com/spf13/cobra"
)

// listCmd represents the parent list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List AWS resources",
	Long:  `List various AWS resources like EC2 instances and AMIs.`,
}

func init() {
	rootCmd.AddCommand(listCmd)
}
