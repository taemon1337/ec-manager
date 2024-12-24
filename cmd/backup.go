package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup an EC2 instance",
	Long:  `Create an AMI backup of an EC2 instance`,
	RunE:  runBackup,
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().StringVarP(&instanceID, "instance", "i", "", "Instance ID to backup")
	backupCmd.MarkFlagRequired("instance")
}

func runBackup(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Initialize AWS clients
	amiService, err := initAWSClients(ctx)
	if err != nil {
		return fmt.Errorf("init AWS clients: %w", err)
	}

	// Create backup
	err = amiService.BackupInstance(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("backup instance: %w", err)
	}

	return nil
}
