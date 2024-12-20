package cmd

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
)

var (
	snapshotID string
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore volumes from snapshots",
	Long: `Restore volumes from snapshots to an instance.
Requires the snapshot ID and instance ID to restore to.`,
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

		// Load AWS configuration
		cfg, err := config.LoadDefaultConfig(cmd.Context())
		if err != nil {
			return fmt.Errorf("unable to load SDK config: %w", err)
		}

		// Create EC2 client
		ec2Client := ec2.NewFromConfig(cfg)

		// Create AMI service
		amiService := ami.NewService(ec2Client)

		fmt.Printf("Starting restore of snapshot %s to instance %s\n", snapshotID, instanceID)
		if err := amiService.RestoreInstance(cmd.Context(), instanceID, snapshotID); err != nil {
			return fmt.Errorf("failed to restore instance: %w", err)
		}

		fmt.Println("Restore completed successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().StringVar(&snapshotID, "snapshot-id", "", "ID of snapshot to restore from")
	restoreCmd.Flags().String("instance-id", "", "ID of instance to restore to")
}
