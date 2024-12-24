package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// AwsCmd represents the aws command
var AwsCmd = &cobra.Command{
	Use:   "aws",
	Short: "AWS configuration",
	Long:  `Configure AWS credentials and settings`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not implemented")
	},
}

func init() {
	rootCmd.AddCommand(AwsCmd)
}
