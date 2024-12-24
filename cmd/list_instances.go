package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
)

// listInstancesCmd represents the list instances command
var listInstancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "List EC2 instances",
	Long:  `List all EC2 instances in your AWS account.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		amiService := ami.NewService(awsClient.GetEC2Client())

		output, err := amiService.DescribeInstances(ctx)
		if err != nil {
			return fmt.Errorf("failed to list instances: %w", err)
		}

		if len(output.Reservations) == 0 {
			fmt.Println("No instances found")
			return nil
		}

		for _, reservation := range output.Reservations {
			for _, instance := range reservation.Instances {
				fmt.Printf("Instance ID: %s\n", *instance.InstanceId)
				if instance.State != nil {
					fmt.Printf("  State: %s\n", instance.State.Name)
				}
				if instance.InstanceType != "" {
					fmt.Printf("  Type: %s\n", instance.InstanceType)
				}
				if instance.PublicIpAddress != nil {
					fmt.Printf("  Public IP: %s\n", *instance.PublicIpAddress)
				}
				if instance.PrivateIpAddress != nil {
					fmt.Printf("  Private IP: %s\n", *instance.PrivateIpAddress)
				}
				fmt.Println()
			}
		}

		return nil
	},
}

func init() {
	listCmd.AddCommand(listInstancesCmd)
}
