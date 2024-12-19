package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
	"github.com/taemon1337/ami-migrate/pkg/ami"
)

var (
	migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Migrate instances to a new AMI",
		Long: `Migrate EC2 instances to a new AMI version.
If --instance-id is provided, migrates that specific instance.
Otherwise, looks for instances with appropriate tags:
- Running instances require both ami-migrate=enabled and ami-migrate-if-running=enabled tags
- Stopped instances only require ami-migrate=enabled tag.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if newAMI == "" {
				return fmt.Errorf("--new-ami is required")
			}

			// Create AMI service
			amiService := ami.NewService(ec2Client)

			// Get current latest AMI
			oldAMI, err := amiService.GetAMIWithTag(context.Background(), "Status", latestTag)
			if err != nil {
				return fmt.Errorf("Failed to get current AMI: %v", err)
			}
			if oldAMI == "" {
				oldAMI = newAMI // If no AMI is marked as latest, use the new AMI as the old one
			}

			// Update AMI tags
			if err := amiService.TagAMI(context.Background(), oldAMI, "Status", "previous"); err != nil {
				log.Printf("Warning: Failed to update old AMI tags: %v", err)
			}
			if err := amiService.TagAMI(context.Background(), newAMI, "Status", latestTag); err != nil {
				log.Printf("Warning: Failed to update new AMI tags: %v", err)
			}

			// Migrate instances
			if instanceID != "" {
				fmt.Printf("Starting migration of instance %s to AMI %s\n", instanceID, newAMI)
				if err := amiService.MigrateInstance(context.Background(), instanceID, oldAMI, newAMI); err != nil {
					return fmt.Errorf("Failed to migrate instance: %v", err)
				}
			} else {
				fmt.Printf("Starting migration from AMI %s to %s\n", oldAMI, newAMI)
				fmt.Printf("Will migrate instances with tag 'ami-migrate=%s'\n", enabledValue)
				fmt.Printf("Instances with 'ami-migrate-if-running=enabled' will be started if needed\n")

				if err := amiService.MigrateInstances(context.Background(), oldAMI, newAMI, enabledValue); err != nil {
					return fmt.Errorf("Failed to migrate instances: %v", err)
				}
			}

			fmt.Println("Migration completed successfully")
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().StringVar(&newAMI, "new-ami", "", "ID of new AMI to migrate to")
	migrateCmd.Flags().StringVar(&instanceID, "instance-id", "", "ID of specific instance to migrate (bypasses tag requirements)")

	// Only initialize AWS client if not already set (for testing)
	if ec2Client == nil {
		// Initialize AWS client
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			log.Fatalf("Unable to load SDK config: %v", err)
		}
		ec2Client = ec2.NewFromConfig(cfg)
	}
}
