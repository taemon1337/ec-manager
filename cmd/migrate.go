package cmd

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/types"
)

// Variables to store flag values
var (
	migrateInstanceID string
	targetAMI        string
	targetVersion    string
	enabled          bool
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate an instance to a new AMI",
	Long:  `Migrate an instance by creating a new instance with the specified AMI and copying over the volumes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate flags
		if migrateInstanceID == "" && !enabled {
			return fmt.Errorf("either --instance or --enabled flag must be set")
		}

		if targetAMI == "" && targetVersion == "" {
			return fmt.Errorf("either --new-ami or --version flag must be specified")
		}

		// Get EC2 client from context
		ec2Client, ok := cmd.Context().Value(types.EC2ClientKey).(types.EC2Client)
		if !ok {
			return fmt.Errorf("failed to get EC2 client")
		}

		if migrateInstanceID != "" {
			_, err := ec2Client.DescribeInstances(cmd.Context(), &ec2.DescribeInstancesInput{
				InstanceIds: []string{migrateInstanceID},
			})
			if err != nil {
				return fmt.Errorf("failed to describe instance: %v", err)
			}
		}

		ctx := cmd.Context()
		amiService := ami.NewService(ec2Client)

		if targetVersion != "" {
			// Get instance OS
			os, err := amiService.GetInstanceOS(ctx, migrateInstanceID)
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

		newInstanceID, err := amiService.MigrateInstance(ctx, migrateInstanceID, targetAMI)
		if err != nil {
			return err
		}

		fmt.Printf("Successfully migrated instance %s to %s (new instance ID: %s)\n", migrateInstanceID, targetAMI, newInstanceID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().StringVarP(&migrateInstanceID, "instance", "i", "", "Instance to migrate")
	migrateCmd.Flags().StringVarP(&targetAMI, "new-ami", "a", "", "New AMI ID to migrate to")
	migrateCmd.Flags().BoolVarP(&enabled, "enabled", "e", false, "Migrate all enabled instances")
	migrateCmd.Flags().StringVarP(&targetVersion, "version", "v", "", "Version to migrate to")
}
