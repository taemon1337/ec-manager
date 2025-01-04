package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/types"
)

// backupCmd represents the backup command
func NewBackupCmd() *cobra.Command {
	var backupInstanceID string

	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup an EC2 instance",
		Long:  `Create a backup AMI of an EC2 instance.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get EC2 client from context
			ec2Client, ok := cmd.Context().Value(types.EC2ClientKey).(types.EC2Client)
			if !ok {
				return fmt.Errorf("failed to get EC2 client")
			}

			amiService := ami.NewService(ec2Client)
			ctx := cmd.Context()

			// Get instance OS for tagging
			os, err := amiService.GetInstanceOS(ctx, backupInstanceID)
			if err != nil {
				return fmt.Errorf("failed to get instance OS: %w", err)
			}

			// Create backup AMI
			amiID, err := amiService.BackupInstance(ctx, backupInstanceID)
			if err != nil {
				return fmt.Errorf("failed to create backup AMI: %w", err)
			}

			// Tag the AMI with OS and version info
			err = amiService.UpdateAMITags(ctx, amiID, map[string]string{
				"OS":         os,
				"BackupType": "manual",
			})
			if err != nil {
				return fmt.Errorf("failed to tag backup AMI: %w", err)
			}

			fmt.Printf("Successfully created backup AMI %s for instance %s\n", amiID, backupInstanceID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&backupInstanceID, "instance-id", "i", "", "Instance ID to backup")
	if err := cmd.MarkFlagRequired("instance-id"); err != nil {
		panic(err)
	}

	return cmd
}

func init() {
	rootCmd.AddCommand(NewBackupCmd())
}
