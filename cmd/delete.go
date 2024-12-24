package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/client"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an EC2 instance",
	Long: `Delete an EC2 instance by its instance ID.
	
	Example:
	  ecman delete --instance-id i-1234567890abcdef0 --user your-user-id`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get instance ID from flag
		instanceID, err := cmd.Flags().GetString("instance-id")
		if err != nil {
			return fmt.Errorf("get instance-id flag: %w", err)
		}

		// Get user ID from flag
		userID, err := cmd.Flags().GetString("user")
		if err != nil {
			return fmt.Errorf("get user flag: %w", err)
		}

		// Get EC2 client
		c := client.NewClient()
		ec2Client, err := c.GetEC2Client(cmd.Context())
		if err != nil {
			return fmt.Errorf("get EC2 client: %w", err)
		}

		// Create AMI service
		svc := ami.NewService(ec2Client)

		// Delete instance
		if err := svc.DeleteInstance(cmd.Context(), userID, instanceID); err != nil {
			return fmt.Errorf("delete instance: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Instance %s deleted successfully\n", instanceID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().String("instance-id", "", "ID of the instance to delete")
	deleteCmd.Flags().String("user", "", "Your user ID")
	deleteCmd.MarkFlagRequired("instance-id")
	deleteCmd.MarkFlagRequired("user")
}
