package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
)

var (
	backupInstanceID string
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup an EC2 instance",
	Long:  `Create a backup AMI of an EC2 instance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		amiService := ami.NewService(awsClient.GetEC2Client())

		// Get instance OS for tagging
		os, err := amiService.GetInstanceOS(ctx, backupInstanceID)
		if err != nil {
			return fmt.Errorf("failed to get instance OS: %w", err)
		}

		// Create backup AMI
		amiID, err := amiService.BackupInstance(ctx, backupInstanceID)
		if err != nil {
			return fmt.Errorf("failed to create backup AMI: %w", err)
		}

		// Tag the AMI with OS and version info
		err = amiService.UpdateAMITags(ctx, amiID, map[string]string{
			"OS":         os,
			"BackupType": "manual",
		})
		if err != nil {
			return fmt.Errorf("failed to tag backup AMI: %w", err)
		}

		fmt.Printf("Successfully created backup AMI %s for instance %s\n", amiID, backupInstanceID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)

	backupCmd.Flags().StringVarP(&backupInstanceID, "instance", "i", "", "Instance ID to backup")
	if err := backupCmd.MarkFlagRequired("instance"); err != nil {
		panic(err)
	}
}
