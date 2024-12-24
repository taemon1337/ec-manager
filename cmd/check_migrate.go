package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/client"
)

var (
	checkMigrationCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Check if your instances need to be migrated to newer AMIs",
		Long: `Check if your instances need to be migrated to newer AMIs. This command will:
1. Check if your instances are running on the latest AMI
2. If not, it will show you which instances need to be migrated
3. You can then use the 'migrate' command to perform the actual migration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get user ID from flag
			userID, err := cmd.Flags().GetString("user")
			if err != nil {
				return fmt.Errorf("get user flag: %w", err)
			}
			if userID == "" {
				return fmt.Errorf("--user flag is required")
			}

			// Get EC2 client
			c := client.NewClient()
			ec2Client, err := c.GetEC2Client(cmd.Context())
			if err != nil {
				return fmt.Errorf("get EC2 client: %w", err)
			}

			// Create AMI service
			amiService := ami.NewService(ec2Client)

			// Check migration status
			status, err := amiService.CheckMigrationStatus(cmd.Context(), userID)
			if err != nil {
				return fmt.Errorf("check migration status: %w", err)
			}

			// Format and print status
			fmt.Fprint(cmd.OutOrStdout(), status.FormatMigrationStatus())

			return nil
		},
	}
)

func init() {
	checkCmd.AddCommand(checkMigrationCmd)
	checkMigrationCmd.Flags().String("user", "", "AWS user ID")
}
