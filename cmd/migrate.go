package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate an EC2 instance to a new AMI",
	Long:  `Migrate an EC2 instance to a new AMI by creating a new instance with the same configuration`,
	RunE:  runMigrate,
}

var (
	newAMI string
)

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().StringVarP(&instanceID, "instance", "i", "", "Instance ID to migrate")
	migrateCmd.Flags().StringVarP(&newAMI, "ami", "a", "", "New AMI ID to migrate to")

	migrateCmd.MarkFlagRequired("instance")
	migrateCmd.MarkFlagRequired("ami")
}

func runMigrate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Initialize AWS clients
	amiService, err := initAWSClients(ctx)
	if err != nil {
		return fmt.Errorf("init AWS clients: %w", err)
	}

	// Migrate instance
	err = amiService.MigrateInstance(ctx, instanceID, newAMI)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}

	fmt.Printf("Successfully migrated instance %s to AMI %s\n", instanceID, newAMI)
	return nil
}
