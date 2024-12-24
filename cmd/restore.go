package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	snapshotID string
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore an EC2 instance from a backup",
	Long: `Restore an EC2 instance from a backup snapshot.
		
		Example:
		  ecman restore --snapshot snap-1234567890abcdef0 --instance i-1234567890abcdef0`,
	RunE: runRestore,
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	restoreCmd.Flags().StringVarP(&instanceID, "instance", "i", "", "Instance ID to restore")
	restoreCmd.Flags().StringVarP(&snapshotID, "snapshot", "s", "", "Snapshot ID to restore from")

	restoreCmd.MarkFlagRequired("instance")
	restoreCmd.MarkFlagRequired("snapshot")
}

func runRestore(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Initialize AWS clients
	amiService, err := initAWSClients(ctx)
	if err != nil {
		return fmt.Errorf("init AWS clients: %w", err)
	}

	// Restore instance
	fmt.Printf("Starting restore of snapshot %s to instance %s\n", snapshotID, instanceID)
	if err := amiService.RestoreInstance(ctx, instanceID, snapshotID); err != nil {
		return fmt.Errorf("failed to restore instance: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Instance %s restored from snapshot %s\n", instanceID, snapshotID)
	return nil
}
