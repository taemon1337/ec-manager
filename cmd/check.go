package cmd

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
	"github.com/taemon1337/ami-migrate/pkg/ami"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if instances need migration",
	Long: `Check if your EC2 instances need to be migrated to a newer AMI.
Shows current and latest AMI details for each instance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, err := cmd.Flags().GetString("user")
		if err != nil {
			return fmt.Errorf("get user flag: %w", err)
		}

		if userID == "" {
			return fmt.Errorf("--user flag is required")
		}

		// Load AWS configuration
		cfg, err := config.LoadDefaultConfig(cmd.Context())
		if err != nil {
			return fmt.Errorf("load AWS config: %w", err)
		}

		// Create EC2 client and AMI service
		ec2Client := ec2.NewFromConfig(cfg)
		amiService := ami.NewService(ec2Client)

		status, err := amiService.CheckMigrationStatus(cmd.Context(), userID)
		if err != nil {
			return fmt.Errorf("check migration status: %w", err)
		}

		fmt.Print(status.FormatMigrationStatus())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().String("user", "", "User ID to check instances for")
	checkCmd.MarkFlagRequired("user")
}
