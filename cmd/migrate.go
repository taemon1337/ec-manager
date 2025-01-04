package cmd

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/client"
	"github.com/taemon1337/ec-manager/pkg/types"
)

// NewMigrateCmd creates a new migrate command
func NewMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate an instance to a new AMI",
		Long:  `Migrate an instance by creating a new instance with the specified AMI and copying over the volumes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get flag values
			instanceID, _ := cmd.Flags().GetString("instance-id")
			targetAMI, _ := cmd.Flags().GetString("new-ami")
			enabled, _ := cmd.Flags().GetBool("enabled")
			targetVersion, _ := cmd.Flags().GetString("version")

			// Validate flags
			if instanceID == "" && !enabled {
				return fmt.Errorf("either --instance-id or --enabled flag must be set")
			}

			if targetAMI == "" && targetVersion == "" {
				return fmt.Errorf("either --new-ami or --version flag must be specified")
			}

			// Get EC2 client from context
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

			if instanceID != "" {
				_, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
					InstanceIds: []string{instanceID},
				})
				if err != nil {
					return fmt.Errorf("failed to describe instance: %v", err)
				}
			}

			amiService := ami.NewService(ec2Client)

			if targetVersion != "" {
				// Get instance OS
				os, err := amiService.GetInstanceOS(ctx, instanceID)
				if err != nil {
					return err
				}

				// Get AMI by version
				ami, err := amiService.GetAMIByVersion(ctx, os, targetVersion)
				if err != nil {
					return err
				}
				targetAMI = *ami.ImageId
			}

			newInstanceID, err := amiService.MigrateInstance(ctx, instanceID, targetAMI)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully migrated instance %s to %s (new instance ID: %s)\n", instanceID, targetAMI, newInstanceID)
			return nil
		},
	}

	cmd.Flags().StringP("instance-id", "i", "", "Instance to migrate")
	cmd.Flags().StringP("new-ami", "a", "", "New AMI ID to migrate to")
	cmd.Flags().BoolP("enabled", "e", false, "Migrate all enabled instances")
	cmd.Flags().StringP("version", "v", "", "Version to migrate to")

	return cmd
}

func init() {
	rootCmd.AddCommand(NewMigrateCmd())
}
