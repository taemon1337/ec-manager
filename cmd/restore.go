package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
)

// RestoreCmd represents the restore command
var RestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore an instance from a snapshot",
	Long:  "Restore an instance by creating and attaching a volume from a snapshot",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		amiService := ami.NewService(awsClient.GetEC2Client())
		return amiService.RestoreInstance(ctx, restoreInstanceID, snapshotID)
	},
}

var (
	restoreInstanceID string
	snapshotID        string
)

func init() {
	rootCmd.AddCommand(RestoreCmd)

	RestoreCmd.Flags().StringVarP(&restoreInstanceID, "instance", "i", "", "Instance ID to restore")
	RestoreCmd.Flags().StringVarP(&snapshotID, "snapshot", "s", "", "Snapshot ID to restore from")

	RestoreCmd.MarkFlagRequired("instance")
	RestoreCmd.MarkFlagRequired("snapshot")
}
