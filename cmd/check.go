package cmd

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check instance migration status",
	Long: `Check if your instances need to be migrated to a newer AMI.
Shows:
- Current AMI details
- Latest available AMI
- Migration recommendation`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get user ID
		userID, err := getUserID(cmd)
		if err != nil {
			return err
		}

		// Load AWS configuration
		cfg, err := config.LoadDefaultConfig(cmd.Context())
		if err != nil {
			return fmt.Errorf("load AWS config: %w", err)
		}

		// Create EC2 client and AMI service
		ec2Client := ec2.NewFromConfig(cfg)
		svc := ami.NewService(ec2Client)

		// Check migration status
		status, err := svc.CheckMigrationStatus(cmd.Context(), userID)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %v", err)
		}

		// Display results
		fmt.Printf("Instance Status for %s:\n", status.InstanceID)
		fmt.Printf("  OS Type:        %s\n", status.OSType)
		fmt.Printf("  Instance Type:  %s\n", status.InstanceType)
		fmt.Printf("  State:          %s\n", status.InstanceState)
		fmt.Printf("  Launch Time:    %s\n", status.LaunchTime.Format(time.RFC3339))
		if status.PrivateIP != "" {
			fmt.Printf("  Private IP:     %s\n", status.PrivateIP)
		}
		if status.PublicIP != "" {
			fmt.Printf("  Public IP:      %s\n", status.PublicIP)
		}

		fmt.Println("\nAMI Status:")
		fmt.Printf("  Current AMI:    %s\n", status.CurrentAMI)
		if status.CurrentAMIInfo != nil {
			fmt.Printf("    Name:         %s\n", status.CurrentAMIInfo.Name)
			fmt.Printf("    Created:      %s\n", status.CurrentAMIInfo.CreatedDate)
		}
		if status.LatestAMI != "" {
			fmt.Printf("  Latest AMI:     %s\n", status.LatestAMI)
			if status.LatestAMIInfo != nil {
				fmt.Printf("    Name:         %s\n", status.LatestAMIInfo.Name)
				fmt.Printf("    Created:      %s\n", status.LatestAMIInfo.CreatedDate)
			}
		}

		fmt.Printf("\nMigration Needed: %v\n", status.NeedsMigration)

		if status.NeedsMigration {
			fmt.Println("\nRun 'ami-migrate migrate' to update your instance to the latest AMI.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().String("user", "", "User ID to check instances for")
	checkCmd.MarkFlagRequired("user")
}
