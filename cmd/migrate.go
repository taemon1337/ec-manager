package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ami-migrate/pkg/ami"
	"github.com/taemon1337/ami-migrate/pkg/client"
)

var (
	migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Migrate EC2 instance to a new AMI",
		Long: `Migrate EC2 instance to a new AMI.
	
	This command will:
	1. Create a new AMI from the instance
	2. Launch a new instance from the AMI
	3. Tag the new instance with the same tags as the old instance`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get flags
			instanceID, err := cmd.Flags().GetString("instance-id")
			if err != nil {
				return fmt.Errorf("get instance-id flag: %w", err)
			}

			newAMI, err := cmd.Flags().GetString("new-ami")
			if err != nil {
				return fmt.Errorf("get new-ami flag: %w", err)
			}

			if newAMI == "" {
				return fmt.Errorf("--new-ami is required")
			}

			// Create AMI service with EC2 client
			amiService := ami.NewService(client.GetEC2Client())

			// Migrate instance
			fmt.Printf("Starting migration of instance %s\n", instanceID)
			err = amiService.MigrateInstance(cmd.Context(), instanceID)
			if err != nil {
				return fmt.Errorf("Failed to migrate instance: %w", err)
			}

			fmt.Printf("Migration completed successfully\n")
			return nil
		},
	}

	instanceID string
	newAMI     string
)

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().StringVar(&instanceID, "instance-id", "", "ID of the instance to migrate")
	migrateCmd.Flags().StringVar(&newAMI, "new-ami", "", "ID of the new AMI to migrate to")
	migrateCmd.MarkFlagRequired("instance-id")
	migrateCmd.MarkFlagRequired("new-ami")
}
