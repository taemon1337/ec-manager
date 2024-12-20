package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ami-migrate/pkg/ami"
	"github.com/taemon1337/ami-migrate/pkg/client"
	"github.com/taemon1337/ami-migrate/pkg/logger"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate EC2 instances to a new AMI",
	Long: `migrate moves EC2 instances to a new AMI. You can specify a single instance
using the --instance-id flag, or migrate all instances with the ami-migrate=enabled tag
by using the --enabled flag. The --new-ami flag is required to specify the target AMI.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate required flags
		instanceID, _ := cmd.Flags().GetString("instance-id")
		enabled, _ := cmd.Flags().GetBool("enabled")
		newAMI, _ := cmd.Flags().GetString("new-ami")

		if instanceID == "" && !enabled {
			return fmt.Errorf("either --instance-id or --enabled flag must be specified")
		}

		if newAMI == "" {
			return fmt.Errorf("--new-ami flag must be specified")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Starting migration process")

		// Get flag values
		instanceID, _ := cmd.Flags().GetString("instance-id")
		enabled, _ := cmd.Flags().GetBool("enabled")
		newAMI, _ := cmd.Flags().GetString("new-ami")

		// Create AWS clients
		ec2Client := client.GetEC2Client()

		// Create AMI service
		svc := ami.NewService(ec2Client)

		// Get instances to migrate
		var instances []string
		if instanceID != "" {
			instances = []string{instanceID}
		} else if enabled {
			// Get all instances with ami-migrate=enabled tag
			taggedInstances, err := svc.ListUserInstances(context.Background(), "ami-migrate")
			if err != nil {
				return fmt.Errorf("failed to list instances: %v", err)
			}
			for _, instance := range taggedInstances {
				instances = append(instances, instance.InstanceID)
			}
		}

		if len(instances) == 0 {
			return fmt.Errorf("no instances found to migrate")
		}

		// Migrate each instance
		for _, instance := range instances {
			if err := svc.MigrateInstance(context.Background(), instance, newAMI); err != nil {
				return fmt.Errorf("failed to migrate instance %s: %v", instance, err)
			}
			logger.Info("Successfully migrated instance", "instanceID", instance)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	// Add flags
	migrateCmd.Flags().String("instance-id", "", "ID of the instance to migrate")
	migrateCmd.Flags().String("new-ami", "", "ID of the new AMI to migrate to")
	migrateCmd.Flags().Bool("enabled", false, "Migrate all instances with ami-migrate=enabled tag")
}
