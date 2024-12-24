package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var checkMigrateCmd = &cobra.Command{
	Use:   "check-migrate",
	Short: "Check if an instance can be migrated",
	Long:  `Check if an EC2 instance can be migrated to a new AMI`,
	RunE:  runCheckMigrate,
}

func init() {
	rootCmd.AddCommand(checkMigrateCmd)

	checkMigrateCmd.Flags().StringVarP(&instanceID, "instance", "i", "", "Instance ID to check")
	checkMigrateCmd.MarkFlagRequired("instance")
}

func runCheckMigrate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Initialize AWS clients
	amiService, err := initAWSClients(ctx)
	if err != nil {
		return fmt.Errorf("init AWS clients: %w", err)
	}

	// Get instance status
	instances, err := amiService.ListUserInstances(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("get instance status: %w", err)
	}

	// Check if instance exists and can be migrated
	if len(instances) == 0 {
		return fmt.Errorf("instance %s not found", instanceID)
	}

	instance := instances[0]
	if instance.State == "running" {
		fmt.Printf("Instance %s can be migrated\n", instanceID)
	} else {
		fmt.Printf("Instance %s cannot be migrated (state: %s)\n", instanceID, instance.State)
	}

	return nil
}
