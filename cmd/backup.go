package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ami-migrate/pkg/ami"
	"github.com/taemon1337/ami-migrate/pkg/client"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup EC2 instances",
	Long: `Backup EC2 instances by creating AMIs.
If --instance-id is provided, backs up that specific instance.
Otherwise, looks for instances with appropriate tags:
- Running instances require both ami-migrate=enabled and ami-migrate-if-running=enabled tags
- Stopped instances only require ami-migrate=enabled tag.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create AMI service with EC2 client
		amiService := ami.NewService(client.GetEC2Client())

		// Get flags
		instanceID, err := cmd.Flags().GetString("instance-id")
		if err != nil {
			return fmt.Errorf("get instance-id flag: %w", err)
		}

		enabledValue, err := cmd.Flags().GetString("enabled-value")
		if err != nil {
			return fmt.Errorf("get enabled-value flag: %w", err)
		}

		// Backup instances
		if instanceID != "" {
			fmt.Printf("Starting backup of instance %s\n", instanceID)
			if err := amiService.BackupInstance(cmd.Context(), instanceID); err != nil {
				return fmt.Errorf("Failed to backup instance: %v", err)
			}
		} else {
			fmt.Printf("Starting backup of instances with tag 'ami-migrate=%s'\n", enabledValue)
			fmt.Printf("Instances with 'ami-migrate-if-running=enabled' will be started if needed\n")

			if err := amiService.BackupInstances(cmd.Context(), enabledValue); err != nil {
				return fmt.Errorf("Failed to backup instances: %v", err)
			}
		}

		fmt.Println("Backup completed successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().String("instance-id", "", "ID of specific instance to backup")
	backupCmd.Flags().String("enabled-value", "enabled", "Value to match for the ami-migrate tag")
}
