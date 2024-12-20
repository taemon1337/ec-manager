package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ami-migrate",
	Short: "AMI Migration Tool",
	Long: `AMI Migration Tool helps you manage EC2 instances and their AMIs.
It supports creating, listing, checking, and migrating instances to new AMIs.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
