package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/client"
	"github.com/taemon1337/ec-manager/pkg/types"
)

// SetupTestCommand is a helper function for setting up test commands
func SetupTestCommand(t *testing.T, use string, setupMock func(*types.MockEC2Client)) *cobra.Command {
	// Initialize logger
	InitTestLogger()

	// Create mock client
	mockClient := types.NewMockEC2Client()
	if setupMock != nil {
		setupMock(mockClient)
	}
	client.SetMockMode(true)
	client.SetMockEC2Client(mockClient)

	// Create a new command based on use
	var cmd *cobra.Command
	switch use {
	case "migrate":
		cmd = &cobra.Command{
			Use:   "migrate",
			Short: "Migrate EC2 instances to new AMIs",
			RunE: func(cmd *cobra.Command, args []string) error {
				instanceID, _ := cmd.Flags().GetString("instance-id")
				enabled, _ := cmd.Flags().GetBool("enabled")
				timeout, _ := cmd.Flags().GetDuration("timeout")

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				ec2Client, err := client.NewClient().GetEC2Client(ctx)
				if err != nil {
					return err
				}
				amiService := ami.NewService(ec2Client)

				if instanceID != "" {
					return amiService.MigrateInstance(ctx, instanceID, "latest")
				}

				if enabled {
					return amiService.MigrateInstances(ctx, "enabled")
				}

				return nil
			},
		}
	case "backup":
		cmd = &cobra.Command{
			Use:   "backup",
			Short: "Create AMI backups of EC2 instances",
			RunE: func(cmd *cobra.Command, args []string) error {
				instanceID, _ := cmd.Flags().GetString("instance-id")
				enabled, _ := cmd.Flags().GetBool("enabled")
				timeout, _ := cmd.Flags().GetDuration("timeout")

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				ec2Client, err := client.NewClient().GetEC2Client(ctx)
				if err != nil {
					return err
				}
				amiService := ami.NewService(ec2Client)

				if instanceID != "" {
					return amiService.BackupInstance(ctx, instanceID)
				}

				if enabled {
					return amiService.BackupInstances(ctx, "enabled")
				}

				return nil
			},
		}
	default:
		cmd = &cobra.Command{
			Use: use,
			RunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
		}
	}

	// Add common flags
	cmd.Flags().String("instance-id", "", "ID of the instance")
	cmd.Flags().Bool("enabled", false, "Process all instances with ami-migrate=enabled tag")
	cmd.Flags().String("log-level", "info", "Log level (debug, info, warn, error)")
	cmd.Flags().Duration("timeout", 10*time.Second, "Timeout for AWS operations")

	return cmd
}

// NewMigrateCmd creates a new migrate command
func NewMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate EC2 instances to new AMIs",
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceID, _ := cmd.Flags().GetString("instance-id")
			enabled, _ := cmd.Flags().GetBool("enabled")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			if instanceID == "" && !enabled {
				return fmt.Errorf("either --instance-id or --enabled flag must be set")
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()

			ec2Client, err := client.NewClient().GetEC2Client(ctx)
			if err != nil {
				return err
			}
			amiService := ami.NewService(ec2Client)

			if instanceID != "" {
				return amiService.MigrateInstance(ctx, instanceID, "latest")
			}

			if enabled {
				return amiService.MigrateInstances(ctx, "enabled")
			}

			return nil
		},
	}
	return cmd
}

// NewBackupCmd creates a new backup command
func NewBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Create AMI backups of EC2 instances",
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceID, _ := cmd.Flags().GetString("instance-id")
			enabled, _ := cmd.Flags().GetBool("enabled")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()

			ec2Client, err := client.NewClient().GetEC2Client(ctx)
			if err != nil {
				return err
			}
			amiService := ami.NewService(ec2Client)

			if instanceID != "" {
				return amiService.BackupInstance(ctx, instanceID)
			}

			if enabled {
				return amiService.BackupInstances(ctx, "enabled")
			}

			return nil
		},
	}
	return cmd
}
