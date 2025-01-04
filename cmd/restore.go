package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/client"
	"github.com/taemon1337/ec-manager/pkg/mock"
	"github.com/taemon1337/ec-manager/pkg/types"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore an instance from a backup",
	Long:  `Restore an instance by creating and attaching a volume from a snapshot, or by using a specific AMI version.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get EC2 client from context
		ctx := cmd.Context()
		ec2Client, ok := ctx.Value(mock.EC2ClientKey).(types.EC2Client)
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

		if restoreVersion != "" {
			// Get instance OS
			os, err := amiService.GetInstanceOS(ctx, restoreInstanceID)
			if err != nil {
				return fmt.Errorf("failed to get instance OS: %w", err)
			}

			// Get AMI by version
			ami, err := amiService.GetAMIByVersion(ctx, os, restoreVersion)
			if err != nil {
				return fmt.Errorf("failed to get AMI version %s: %w", restoreVersion, err)
			}

			// Migrate to this version
			newInstanceID, err := amiService.MigrateInstance(ctx, restoreInstanceID, *ami.ImageId)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully restored instance %s to version %s (new instance: %s)\n",
				restoreInstanceID, restoreVersion, newInstanceID)
			return nil
		}

		if snapshotID == "" {
			return fmt.Errorf("either --snapshot or --version must be specified")
		}

		// Restore from snapshot
		err := amiService.RestoreInstance(ctx, restoreInstanceID, snapshotID)
		if err != nil {
			return err
		}

		fmt.Printf("Successfully restored instance %s from snapshot %s\n", restoreInstanceID, snapshotID)
		return nil
	},
}

var (
	restoreInstanceID string
	snapshotID        string
	restoreVersion    string
)

func init() {
	rootCmd.AddCommand(restoreCmd)

	restoreCmd.Flags().StringVarP(&restoreInstanceID, "instance-id", "i", "", "Instance ID to restore")
	restoreCmd.Flags().StringVarP(&snapshotID, "snapshot", "s", "", "Snapshot ID to restore from (optional if using --version)")
	restoreCmd.Flags().StringVarP(&restoreVersion, "version", "v", "", "Version to restore to (optional if using --snapshot)")

	if err := restoreCmd.MarkFlagRequired("instance-id"); err != nil {
		panic(err)
	}
}
