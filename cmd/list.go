package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
)

// ListCmd represents the list command
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List EC2 instances",
	Long:  "List all EC2 instances in the account",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		amiService := ami.NewService(awsClient.GetEC2Client())
		
		output, err := amiService.DescribeInstances(ctx)
		if err != nil {
			return fmt.Errorf("failed to list instances: %w", err)
		}

		for _, reservation := range output.Reservations {
			for _, instance := range reservation.Instances {
				fmt.Printf("Instance ID: %s\n", *instance.InstanceId)
				fmt.Printf("  State: %s\n", instance.State.Name)
				fmt.Printf("  Instance Type: %s\n", instance.InstanceType)
				fmt.Printf("  Launch Time: %s\n", instance.LaunchTime)
				fmt.Println()
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(ListCmd)
}
