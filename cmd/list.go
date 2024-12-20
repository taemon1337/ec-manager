package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/taemon1337/ami-migrate/pkg/ami"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List your EC2 instances",
	Long: `List all EC2 instances owned by you.
Shows instance details including:
- Name and ID
- OS type and size
- Current state
- IP addresses
- AMI status (current and if migration is needed)`,
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

		instances, err := amiService.ListUserInstances(cmd.Context(), userID)
		if err != nil {
			return fmt.Errorf("list instances: %w", err)
		}

		if len(instances) == 0 {
			fmt.Printf("No instances found for user: %s\n", userID)
			fmt.Println("\nTo create a new instance, run:")
			fmt.Printf("  ami-migrate create --user %s\n", userID)
			return nil
		}

		fmt.Printf("Found %d instance(s):\n\n", len(instances))
		for _, instance := range instances {
			fmt.Print(instance.FormatInstanceSummary())
			fmt.Println()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().String("user", "", "User ID to list instances for")
	listCmd.MarkFlagRequired("user")
}
