package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/client"
	"github.com/taemon1337/ec-manager/pkg/types"
)

// NewCheckMigrateCmd creates a new check migrate command
func NewCheckMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Check instances that need migration",
		Long:  "Check and list EC2 instances that need to be migrated",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			ec2Client, ok := ctx.Value(types.EC2ClientKey).(types.EC2Client)
			if !ok {
				if awsClient == nil {
					var err error
					awsClient, err = client.NewClient(false, "us-east-1", "default")
					if err != nil {
						return fmt.Errorf("failed to create AWS client: %w", err)
					}
				}
				ec2Client = awsClient.GetEC2Client()
			}
			amiService := ami.NewService(ec2Client)

			instanceID, _ := cmd.Flags().GetString("check-instance-id")
			targetAMI, _ := cmd.Flags().GetString("check-target-ami")

			if instanceID != "" {
				instance, err := amiService.DescribeInstance(ctx, instanceID)
				if err != nil {
					return fmt.Errorf("failed to describe instance: %w", err)
				}
				if instance == nil {
					return fmt.Errorf("instance not found: %s", instanceID)
				}

				if targetAMI != "" {
					images, err := amiService.DescribeImages(ctx, []string{targetAMI})
					if err != nil {
						return fmt.Errorf("failed to describe AMI: %w", err)
					}
					if len(images.Images) == 0 {
						return fmt.Errorf("AMI not found: %s", targetAMI)
					}
				}

				fmt.Printf("Instance ID: %s\n", *instance.InstanceId)
				fmt.Printf("  State: %s\n", instance.State.Name)
				fmt.Printf("  Instance Type: %s\n", instance.InstanceType)
				fmt.Printf("  Launch Time: %s\n", instance.LaunchTime)
				return nil
			}

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

	cmd.Flags().StringP("check-instance-id", "i", "", "Instance ID to check for migration")
	cmd.Flags().StringP("check-target-ami", "a", "", "New AMI ID to migrate to")

	return cmd
}

func init() {
	checkCmd.AddCommand(NewCheckMigrateCmd())
}
