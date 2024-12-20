package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/client"
)

const usage = `Usage: ec-manager <action> [options]

Actions:
  migrate     Migrate EC2 instances to a new AMI
  backup      Create a backup AMI from an EC2 instance
  restore     Restore volumes from snapshots
  list        List your EC2 instances
  check       Check instance status
  delete      Delete an EC2 instance
  help        Show this help message

Options:
  -enabled-value string   Value to match for the ami-migrate tag (default "enabled")
  -timeout duration       Timeout for operations (default 10m)
  -instance-id string     ID of the instance to operate on
  -new-ami string        ID of the new AMI to migrate to (required for migrate)
  -snapshot-id string    ID of the snapshot to restore from (required for restore)

Example:
  ec-manager migrate -enabled-value=enabled -timeout=15m
  ec-manager backup -instance-id=i-1234567890abcdef0
  ec-manager list
  ec-manager help

For more information about a specific action, use: ec-manager help <action>`

func main() {
	rootCmd := &cobra.Command{
		Use:   "ec-manager",
		Short: "EC2 Manager",
	}

	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate EC2 instances to a new AMI",
		Run: func(cmd *cobra.Command, args []string) {
			enabledValue, _ := cmd.Flags().GetString("enabled-value")
			newAMI, _ := cmd.Flags().GetString("new-ami")
			timeoutValue, _ := cmd.Flags().GetDuration("timeout")

			if newAMI == "" {
				log.Fatal("Error: -new-ami is required for migrate action")
			}

			fmt.Printf("Starting migration for instances with tag 'ami-migrate=%s'\n", enabledValue)
			fmt.Printf("Instances with 'ami-migrate-if-running=enabled' will be started if needed\n")

			cfg, err := client.LoadAWSConfig(context.Background())
			if err != nil {
				log.Fatal(err)
			}

			ec2Client := ec2.NewFromConfig(cfg)
			amiService := ami.NewService(ec2Client)

			ctx, cancel := context.WithTimeout(context.Background(), timeoutValue)
			defer cancel()

			if err := amiService.MigrateInstances(ctx, enabledValue); err != nil {
				log.Fatalf("Failed to migrate instances: %v", err)
			}

			fmt.Println("Migration completed successfully")
		},
	}

	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Create a backup AMI from an EC2 instance",
		Run: func(cmd *cobra.Command, args []string) {
			instanceID, _ := cmd.Flags().GetString("instance-id")
			timeoutValue, _ := cmd.Flags().GetDuration("timeout")

			if instanceID == "" {
				log.Fatal("Error: -instance-id is required for backup action")
			}

			fmt.Printf("Starting backup for instance %s\n", instanceID)

			cfg, err := client.LoadAWSConfig(context.Background())
			if err != nil {
				log.Fatal(err)
			}

			ec2Client := ec2.NewFromConfig(cfg)
			amiService := ami.NewService(ec2Client)

			ctx, cancel := context.WithTimeout(context.Background(), timeoutValue)
			defer cancel()

			if err := amiService.BackupInstance(ctx, instanceID); err != nil {
				log.Fatalf("Failed to backup instance: %v", err)
			}

			fmt.Println("Backup completed successfully")
		},
	}

	restoreCmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore volumes from snapshots",
		Run: func(cmd *cobra.Command, args []string) {
			instanceID, _ := cmd.Flags().GetString("instance-id")
			snapshotID, _ := cmd.Flags().GetString("snapshot-id")
			timeoutValue, _ := cmd.Flags().GetDuration("timeout")

			if instanceID == "" || snapshotID == "" {
				log.Fatal("Error: both -instance-id and -snapshot-id are required for restore action")
			}

			fmt.Printf("Starting restore of snapshot %s to instance %s\n", snapshotID, instanceID)

			cfg, err := client.LoadAWSConfig(context.Background())
			if err != nil {
				log.Fatal(err)
			}

			ec2Client := ec2.NewFromConfig(cfg)
			amiService := ami.NewService(ec2Client)

			ctx, cancel := context.WithTimeout(context.Background(), timeoutValue)
			defer cancel()

			if err := amiService.RestoreInstance(ctx, instanceID, snapshotID); err != nil {
				log.Fatalf("Failed to restore instance: %v", err)
			}

			fmt.Println("Restore completed successfully")
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List your EC2 instances",
		Run: func(cmd *cobra.Command, args []string) {
			timeoutValue, _ := cmd.Flags().GetDuration("timeout")

			fmt.Println("Listing instances...")

			cfg, err := client.LoadAWSConfig(context.Background())
			if err != nil {
				log.Fatal(err)
			}

			ec2Client := ec2.NewFromConfig(cfg)
			amiService := ami.NewService(ec2Client)

			ctx, cancel := context.WithTimeout(context.Background(), timeoutValue)
			defer cancel()

			instances, err := amiService.ListUserInstances(ctx, "")
			if err != nil {
				log.Fatalf("Failed to list instances: %v", err)
			}

			for _, instance := range instances {
				fmt.Printf("Instance %s: %s\n", instance.InstanceID, instance.Name)
			}
		},
	}

	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Check instance status",
		Run: func(cmd *cobra.Command, args []string) {
			instanceID, _ := cmd.Flags().GetString("instance-id")
			//timeoutValue, _ := cmd.Flags().GetDuration("timeout")

			if instanceID == "" {
				log.Fatal("Error: -instance-id is required for check action")
			}

			fmt.Printf("Checking instance %s\n", instanceID)

			// TODO: Implement check functionality
		},
	}

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an EC2 instance",
		Run: func(cmd *cobra.Command, args []string) {
			instanceID, _ := cmd.Flags().GetString("instance-id")
			timeoutValue, _ := cmd.Flags().GetDuration("timeout")

			if instanceID == "" {
				log.Fatal("Error: -instance-id is required for delete action")
			}

			fmt.Printf("Deleting instance %s\n", instanceID)

			cfg, err := client.LoadAWSConfig(context.Background())
			if err != nil {
				log.Fatal(err)
			}

			ec2Client := ec2.NewFromConfig(cfg)
			amiService := ami.NewService(ec2Client)

			ctx, cancel := context.WithTimeout(context.Background(), timeoutValue)
			defer cancel()

			if err := amiService.DeleteInstance(ctx, "userID", instanceID); err != nil {
				log.Fatalf("Failed to delete instance: %v", err)
			}

			fmt.Println("Instance deleted successfully")
		},
	}

	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(deleteCmd)

	migrateCmd.Flags().StringP("enabled-value", "e", "enabled", "Value to match for the ami-migrate tag")
	migrateCmd.Flags().StringP("new-ami", "n", "", "ID of the new AMI to migrate to")
	migrateCmd.Flags().DurationP("timeout", "t", 10*time.Minute, "Timeout for operations")

	backupCmd.Flags().StringP("instance-id", "i", "", "ID of the instance to operate on")
	backupCmd.Flags().DurationP("timeout", "t", 10*time.Minute, "Timeout for operations")

	restoreCmd.Flags().StringP("instance-id", "i", "", "ID of the instance to operate on")
	restoreCmd.Flags().StringP("snapshot-id", "s", "", "ID of the snapshot to restore from")
	restoreCmd.Flags().DurationP("timeout", "t", 10*time.Minute, "Timeout for operations")

	listCmd.Flags().DurationP("timeout", "t", 10*time.Minute, "Timeout for operations")

	checkCmd.Flags().StringP("instance-id", "i", "", "ID of the instance to operate on")
	checkCmd.Flags().DurationP("timeout", "t", 10*time.Minute, "Timeout for operations")

	deleteCmd.Flags().StringP("instance-id", "i", "", "ID of the instance to operate on")
	deleteCmd.Flags().DurationP("timeout", "t", 10*time.Minute, "Timeout for operations")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
