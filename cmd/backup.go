package cmd

import (
	"context"

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
		return amiService.BackupInstance(ctx, backupInstanceID)
	},
}

var backupInstanceID string

func init() {
	rootCmd.AddCommand(BackupCmd)

	BackupCmd.Flags().StringVarP(&backupInstanceID, "instance", "i", "", "Instance ID to backup")
	BackupCmd.MarkFlagRequired("instance")
}
