package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/logger"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate an EC2 instance to a new AMI",
	Long: `Migrate an EC2 instance to a new AMI. You can specify a single instance
using the --instance-id flag, or migrate all instances with the ami-migrate=enabled tag
by using the --enabled flag.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate required flags
		instanceID, _ := cmd.Flags().GetString("instance-id")
		enabled, _ := cmd.Flags().GetBool("enabled")
		newAMI, _ := cmd.Flags().GetString("new-ami")

		if instanceID == "" && !enabled {
			return fmt.Errorf("required flag(s) \"instance-id\" or \"enabled\" not set")
		}

		if newAMI == "" {
			return fmt.Errorf("--new-ami flag must be specified")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Starting migration process")

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
		newAMI, _ := cmd.Flags().GetString("new-ami")

		// Check if we're migrating a single instance or all enabled instances
		if instanceID != "" {
			if err := amiService.MigrateInstance(ctx, instanceID, newAMI); err != nil {
				return fmt.Errorf("failed to migrate instance %s: %v", instanceID, err)
			}
			logger.Info("Successfully migrated instance", "instanceID", instanceID)
			return nil
		}

		// Migrate all enabled instances
		if enabled {
			if err := amiService.MigrateInstances(ctx, "enabled"); err != nil {
				return fmt.Errorf("failed to migrate instances: %v", err)
			}
			return nil
		}

		return fmt.Errorf("either --instance-id or --enabled flag must be set")
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	// Add flags
	migrateCmd.Flags().String("instance-id", "", "ID of the instance to migrate")
	migrateCmd.Flags().String("new-ami", "", "ID of the new AMI to migrate to")
	migrateCmd.Flags().Bool("enabled", false, "Migrate all instances with ami-migrate=enabled tag")
}
