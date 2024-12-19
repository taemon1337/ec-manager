package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
	"github.com/taemon1337/ami-migrate/pkg/ami"
)

var (
	snapshotID string
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore volumes from snapshots",
	Long: `Restore volumes from snapshots to an instance.
Requires the snapshot ID and instance ID to restore to.`,
	Run: func(cmd *cobra.Command, args []string) {
		if snapshotID == "" {
			log.Fatal("--snapshot-id is required")
		}
		instanceID, err := cmd.Flags().GetString("instance-id")
		if err != nil {
			log.Fatal(err)
		}
		if instanceID == "" {
			log.Fatal("--instance-id is required")
		}

		// Load AWS configuration
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			log.Fatalf("Unable to load SDK config: %v", err)
		}

		// Create EC2 client
		ec2Client := ec2.NewFromConfig(cfg)

		// Create AMI service
		amiService := ami.NewService(ec2Client)

		fmt.Printf("Starting restore of snapshot %s to instance %s\n", snapshotID, instanceID)
		if err := amiService.RestoreInstance(context.Background(), instanceID, snapshotID); err != nil {
			log.Fatalf("Failed to restore instance: %v", err)
		}

		fmt.Println("Restore completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().StringVar(&snapshotID, "snapshot-id", "", "ID of snapshot to restore from")
	restoreCmd.Flags().String("instance-id", "", "ID of instance to restore to")
}
