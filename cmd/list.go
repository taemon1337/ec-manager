package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/logger"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List your EC2 instances",
	Long: `List all EC2 instances owned by you.
Shows instance details including:
- Instance ID and name
- OS type and size
- Current state
- IP addresses
- Current and latest AMI versions`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get user ID
		userID, err := getUserID(cmd)
		if err != nil {
			return err
		}

		// Get EC2 client from our client package
		ctx := cmd.Context()
		ec2Client, err := awsClient.GetEC2Client(ctx)
		if err != nil {
			return fmt.Errorf("failed to get EC2 client: %w", err)
		}

		// Create AMI service
		amiService := ami.NewService(ec2Client)

		// List instances
		logger.Info("Listing instances for user: " + userID)
		instances, err := amiService.ListUserInstances(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to list instances: %v", err)
		}

		// Display results
		if len(instances) == 0 {
			fmt.Printf("No instances found for user: %s\n", userID)
			fmt.Println("\nTo create a new instance:")
			fmt.Printf("  ami-migrate create --user %s\n", userID)
			return nil
		}

		fmt.Printf("Found %d instance(s):\n\n", len(instances))
		for _, instance := range instances {
			fmt.Printf("Instance: %s (%s)\n", instance.Name, instance.InstanceID)
			fmt.Printf("  OS:           %s\n", instance.OSType)
			fmt.Printf("  Size:         %s\n", instance.Size)
			fmt.Printf("  State:        %s\n", instance.State)
			fmt.Printf("  Launch Time:  %s\n", instance.LaunchTime.Format(time.RFC3339))
			if instance.PrivateIP != "" {
				fmt.Printf("  Private IP:   %s\n", instance.PrivateIP)
			}
			if instance.PublicIP != "" {
				fmt.Printf("  Public IP:    %s\n", instance.PublicIP)
			}
			fmt.Printf("  Current AMI:  %s\n", instance.CurrentAMI)
			if instance.LatestAMI != "" && instance.LatestAMI != instance.CurrentAMI {
				fmt.Printf("  Latest AMI:   %s (migration available)\n", instance.LatestAMI)
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().String("user", "", "User ID to list instances for")
	listCmd.MarkFlagRequired("user")
}
