package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
)

// checkMigrateCmd represents the migrate subcommand of check
var checkMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Check instances that need migration",
	Long:  "Check and list EC2 instances that need to be migrated",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		amiService := ami.NewService(awsClient.GetEC2Client())
		
		output, err := amiService.DescribeInstances(ctx)
		if err != nil {
			return fmt.Errorf("failed to list instances: %w", err)
		}

		fmt.Println("Instances that need migration:")
		for _, reservation := range output.Reservations {
			for _, instance := range reservation.Instances {
				for _, tag := range instance.Tags {
					if *tag.Key == "ami-migrate" && *tag.Value == "enabled" {
						fmt.Printf("Instance ID: %s\n", *instance.InstanceId)
						fmt.Printf("  State: %s\n", instance.State.Name)
						fmt.Printf("  Instance Type: %s\n", instance.InstanceType)
						fmt.Printf("  Launch Time: %s\n", instance.LaunchTime)
						fmt.Println()
					}
				}
			}
		}

		return nil
	},
}

func init() {
	checkCmd.AddCommand(checkMigrateCmd)
}
