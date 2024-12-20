package cmd

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
	"github.com/taemon1337/ami-migrate/pkg/ami"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an EC2 instance",
	Long: `Delete an EC2 instance that you own.
Requires:
- Instance ID
- Your user ID (to verify ownership)

The command will show instance details and ask for confirmation before deletion.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, err := cmd.Flags().GetString("user")
		if err != nil {
			return fmt.Errorf("get user flag: %w", err)
		}

		instanceID, err := cmd.Flags().GetString("instance-id")
		if err != nil {
			return fmt.Errorf("get instance-id flag: %w", err)
		}

		if userID == "" || instanceID == "" {
			return fmt.Errorf("both --user and --instance-id flags are required")
		}

		// Load AWS configuration
		cfg, err := config.LoadDefaultConfig(cmd.Context())
		if err != nil {
			return fmt.Errorf("load AWS config: %w", err)
		}

		// Create EC2 client and AMI service
		ec2Client := ec2.NewFromConfig(cfg)
		amiService := ami.NewService(ec2Client)

		// Get instance details
		instances, err := amiService.ListUserInstances(cmd.Context(), userID)
		if err != nil {
			return fmt.Errorf("list instances: %w", err)
		}

		var instance *ami.InstanceSummary
		for i, inst := range instances {
			if inst.InstanceID == instanceID {
				instance = &instances[i]
				break
			}
		}

		if instance == nil {
			return fmt.Errorf("instance %s not found or not owned by user %s", instanceID, userID)
		}

		// Show instance details and ask for confirmation
		fmt.Printf("\nInstance to delete:\n")
		fmt.Print(instance.FormatInstanceSummary())
		fmt.Print("\nAre you sure you want to delete this instance? [y/N]: ")

		var confirm string
		fmt.Scanln(&confirm)
		if !strings.EqualFold(confirm, "y") {
			fmt.Println("Instance deletion cancelled")
			return nil
		}

		if err := amiService.DeleteInstance(cmd.Context(), userID, instanceID); err != nil {
			return fmt.Errorf("delete instance: %w", err)
		}

		fmt.Println("Instance deleted successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().String("user", "", "Your user ID")
	deleteCmd.Flags().String("instance-id", "", "ID of the instance to delete")
	deleteCmd.MarkFlagRequired("user")
	deleteCmd.MarkFlagRequired("instance-id")
}
