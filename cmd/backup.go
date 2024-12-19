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

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup instance volumes",
	Long: `Create snapshots of all volumes attached to the instance.
If --instance-id is provided, backs up that specific instance.
Otherwise, looks for instances with appropriate tags:
- Running instances require both ami-migrate=enabled and ami-migrate-if-running=enabled tags
- Stopped instances only require ami-migrate=enabled tag.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create AMI service
		amiService := ami.NewService(ec2Client)

		if instanceID != "" {
			fmt.Printf("Starting backup of instance %s\n", instanceID)
			if err := amiService.BackupInstance(context.Background(), instanceID); err != nil {
				log.Fatalf("Failed to backup instance: %v", err)
			}
		} else {
			// Backup instances
			fmt.Printf("Starting backup of instances with tag 'ami-migrate=%s'\n", enabledValue)
			fmt.Printf("Instances with 'ami-migrate-if-running=enabled' will be started if needed\n")

			if err := amiService.BackupInstances(context.Background(), enabledValue); err != nil {
				log.Fatalf("Failed to backup instances: %v", err)
			}
		}

		fmt.Println("Backup completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().StringVar(&instanceID, "instance-id", "", "ID of specific instance to backup (bypasses tag requirements)")

	// Only initialize AWS client if not already set (for testing)
	if ec2Client == nil {
		// Initialize AWS client
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			log.Fatalf("Unable to load SDK config: %v", err)
		}
		ec2Client = ec2.NewFromConfig(cfg)
	}
}
