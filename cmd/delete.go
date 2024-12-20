package cmd

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an EC2 instance",
	Long: `Delete an EC2 instance owned by you.
Requires the instance ID to be specified.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get user ID
		userID, err := getUserID(cmd)
		if err != nil {
			return err
		}

		// Get instance ID
		instanceID, err := cmd.Flags().GetString("instance-id")
		if err != nil {
			return err
		}

		// Validate required flags
		if instanceID == "" {
			return fmt.Errorf("--instance-id flag is required")
		}

		// Load AWS configuration
		cfg, err := config.LoadDefaultConfig(cmd.Context())
		if err != nil {
			return fmt.Errorf("load AWS config: %w", err)
		}

		// Create EC2 client and AMI service
		ec2Client := ec2.NewFromConfig(cfg)
		svc := ami.NewService(ec2Client)

		// Verify instance ownership
		instances, err := svc.ListUserInstances(cmd.Context(), userID)
		if err != nil {
			return fmt.Errorf("failed to list instances: %v", err)
		}

		var found bool
		for _, instance := range instances {
			if instance.InstanceID == instanceID {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("instance %s not found or not owned by user %s", instanceID, userID)
		}

		// Confirm deletion
		fmt.Printf("Are you sure you want to delete instance %s? [y/N] ", instanceID)
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println("Deletion cancelled")
			return nil
		}

		// Delete instance
		if err := svc.DeleteInstance(cmd.Context(), userID, instanceID); err != nil {
			return fmt.Errorf("failed to delete instance: %v", err)
		}

		fmt.Printf("Successfully deleted instance %s\n", instanceID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().String("instance-id", "", "ID of the instance to delete")
	deleteCmd.MarkFlagRequired("instance-id")
}
