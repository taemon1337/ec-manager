package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/client"
)

var (
	snapshotID string
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore an instance from a snapshot",
	Long: `Restore an instance from a snapshot.
		
		Example:
		  ecman restore --snapshot-id snap-1234567890abcdef0 --instance-id i-1234567890abcdef0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if snapshotID == "" {
			return fmt.Errorf("--snapshot-id is required")
		}
		instanceID, err := cmd.Flags().GetString("instance-id")
		if err != nil {
			return fmt.Errorf("failed to get instance-id flag: %w", err)
		}
		if instanceID == "" {
			return fmt.Errorf("--instance-id is required")
		}

		// Get EC2 client
		c := client.NewClient()
		ec2Client, err := c.GetEC2Client(cmd.Context())
		if err != nil {
			return fmt.Errorf("get EC2 client: %w", err)
		}

		// Create AMI service
		amiService := ami.NewService(ec2Client)

		fmt.Printf("Starting restore of snapshot %s to instance %s\n", snapshotID, instanceID)
		if err := amiService.RestoreInstance(cmd.Context(), instanceID, snapshotID); err != nil {
			return fmt.Errorf("failed to restore instance: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Instance %s restored from snapshot %s\n", instanceID, snapshotID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().StringVar(&snapshotID, "snapshot-id", "", "ID of snapshot to restore from")
	restoreCmd.Flags().String("instance-id", "", "ID of instance to restore to")
}
