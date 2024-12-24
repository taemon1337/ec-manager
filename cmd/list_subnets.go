package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
)

// listSubnetsCmd represents the list subnets command
var listSubnetsCmd = &cobra.Command{
	Use:   "subnets",
	Short: "List available VPC subnets",
	Long:  `List all available VPC subnets in your AWS account.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		amiService := ami.NewService(awsClient.GetEC2Client())

		subnets, err := amiService.ListSubnets(ctx)
		if err != nil {
			return fmt.Errorf("failed to list subnets: %w", err)
		}

		if len(subnets) == 0 {
			fmt.Println("No subnets found")
			return nil
		}

		for _, subnet := range subnets {
			fmt.Printf("Subnet ID: %s\n", *subnet.SubnetId)
			fmt.Printf("  VPC ID: %s\n", *subnet.VpcId)
			fmt.Printf("  CIDR Block: %s\n", *subnet.CidrBlock)
			fmt.Printf("  Availability Zone: %s\n", *subnet.AvailabilityZone)
			if subnet.AvailableIpAddressCount != nil {
				fmt.Printf("  Available IPs: %d\n", *subnet.AvailableIpAddressCount)
			}
			if len(subnet.Tags) > 0 {
				fmt.Println("  Tags:")
				for _, tag := range subnet.Tags {
					if tag.Key != nil && tag.Value != nil {
						fmt.Printf("    %s: %s\n", *tag.Key, *tag.Value)
					}
				}
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	listCmd.AddCommand(listSubnetsCmd)
}
