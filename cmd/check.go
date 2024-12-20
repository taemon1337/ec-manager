package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if your instance needs migration",
	Long: `Check if your instance needs to be migrated to a newer AMI.
This command will:
1. Find your instance using the Owner tag
2. Determine its OS type
3. Check if a newer AMI is available
4. Show detailed information about the current and latest AMIs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, err := cmd.Flags().GetString("user")
		if err != nil {
			return fmt.Errorf("get user flag: %w", err)
		}

		if userID == "" {
			return fmt.Errorf("--user flag is required")
		}

		status, err := amiService.CheckMigrationStatus(cmd.Context(), userID)
		if err != nil {
			return fmt.Errorf("check migration status: %w", err)
		}

		// Print the formatted status
		fmt.Println(status.FormatMigrationStatus())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().String("user", "", "Your user ID (required)")
	checkCmd.MarkFlagRequired("user")
}
