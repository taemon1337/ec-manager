package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/logger"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup an EC2 instance",
	Long: `Backup an EC2 instance by creating an AMI. You can specify a single instance
using the --instance-id flag, or backup all instances with the ami-migrate=enabled tag
by using the --enabled flag.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate required flags
		instanceID, _ := cmd.Flags().GetString("instance-id")
		enabled, _ := cmd.Flags().GetBool("enabled")

		if instanceID == "" && !enabled {
			return fmt.Errorf("required flag(s) \"instance-id\" or \"enabled\" not set")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Starting backup process")

		// Create AMI service with AWS client
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		ec2Client, err := awsClient.GetEC2Client(ctx)
		if err != nil {
			return fmt.Errorf("failed to get EC2 client: %w", err)
		}
		amiService := ami.NewService(ec2Client)

		// Get flag values
		instanceID, _ := cmd.Flags().GetString("instance-id")
		enabled, _ := cmd.Flags().GetBool("enabled")

		// Check if we're backing up a single instance or all enabled instances
		if instanceID != "" {
			if err := amiService.BackupInstance(ctx, instanceID); err != nil {
				return fmt.Errorf("failed to backup instance %s: %v", instanceID, err)
			}
			return nil
		}

		// Backup all enabled instances
		if enabled {
			if err := amiService.BackupInstances(ctx, "enabled"); err != nil {
				return fmt.Errorf("failed to backup instances: %v", err)
			}
			return nil
		}

		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)

	// Add flags
	backupCmd.Flags().String("instance-id", "", "ID of the instance")
	backupCmd.Flags().Bool("enabled", false, "Process all instances with ami-migrate=enabled tag")
}
