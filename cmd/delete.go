package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an EC2 instance",
	Long:  `Delete an EC2 instance and its associated resources`,
	RunE:  runDelete,
}

var userID string

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringVarP(&instanceID, "instance", "i", "", "Instance ID to delete")
	deleteCmd.Flags().StringVarP(&userID, "user", "u", "", "User ID owning the instance")
	deleteCmd.MarkFlagRequired("instance")
}

func runDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Initialize AWS clients
	amiService, err := initAWSClients(ctx)
	if err != nil {
		return fmt.Errorf("init AWS clients: %w", err)
	}

	// Delete instance
	err = amiService.DeleteInstance(ctx, userID, instanceID)
	if err != nil {
		return fmt.Errorf("delete instance: %w", err)
	}

	fmt.Printf("Successfully deleted instance %s\n", instanceID)
	return nil
}
