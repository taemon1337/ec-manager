package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/client"
	"github.com/taemon1337/ec-manager/pkg/logger"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a backup AMI from an EC2 instance",
	Long: `backup creates a backup AMI from an EC2 instance. You can specify a single instance
using the --instance-id flag, or back up all instances with the ami-migrate=enabled tag
by using the --enabled flag.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate required flags
		instanceID, _ := cmd.Flags().GetString("instance-id")
		enabled, _ := cmd.Flags().GetBool("enabled")

		if instanceID == "" && !enabled {
			return fmt.Errorf("required flag(s) \"instance-id\" not set")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Starting backup process")

		// Get flag values
		instanceID, _ := cmd.Flags().GetString("instance-id")
		enabled, _ := cmd.Flags().GetBool("enabled")

		// Create AWS clients
		ec2Client, err := client.GetEC2Client(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get EC2 client: %w", err)
		}

		// Create AMI service
		svc := ami.NewService(ec2Client)

		// Get instances to backup
		var instances []string
		if instanceID != "" {
			instances = []string{instanceID}
		} else if enabled {
			// Get all instances with ami-migrate=enabled tag
			taggedInstances, err := svc.ListUserInstances(cmd.Context(), "ami-migrate")
			if err != nil {
				return fmt.Errorf("failed to list instances: %v", err)
			}
			for _, instance := range taggedInstances {
				instances = append(instances, instance.InstanceID)
			}
		}

		if len(instances) == 0 {
			return fmt.Errorf("no instances found to backup")
		}

		// Backup each instance
		for _, instance := range instances {
			logger.Info(fmt.Sprintf("Creating backup AMI for instance %s", instance))
			if err := svc.BackupInstance(cmd.Context(), instance); err != nil {
				return fmt.Errorf("failed to backup instance %s: %v", instance, err)
			}
			logger.Info(fmt.Sprintf("Successfully created backup for instance %s", instance))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)

	// Add flags
	backupCmd.Flags().String("instance-id", "", "ID of the instance to backup")
	backupCmd.Flags().Bool("enabled", false, "Backup all instances with ami-migrate=enabled tag")
}
