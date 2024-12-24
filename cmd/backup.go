package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
)

// BackupCmd represents the backup command
var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup an EC2 instance",
	Long:  "Create a backup of an EC2 instance by creating a snapshot",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		amiService := ami.NewService(awsClient.GetEC2Client())
		
		imageID, err := amiService.BackupInstance(ctx, backupInstanceID)
		if err != nil {
			return err
		}

		fmt.Printf("Created backup AMI %s for instance %s\n", imageID, backupInstanceID)
		return nil
	},
}

var backupInstanceID string

func init() {
	rootCmd.AddCommand(BackupCmd)

	BackupCmd.Flags().StringVarP(&backupInstanceID, "instance", "i", "", "Instance ID to backup")
	BackupCmd.MarkFlagRequired("instance")
}
