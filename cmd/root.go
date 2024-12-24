package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/client"
	"github.com/taemon1337/ec-manager/pkg/config"
	"github.com/taemon1337/ec-manager/pkg/logger"
)

var (
	// Common flags
	mockMode    bool
	logLevel    string
	timeout     time.Duration
	enabledFlag bool
	instanceID  string
	newAMI      string
	userID      string

	// AWS client
	awsClient *client.Client
)

func init() {
	awsClient = client.NewClient()
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ecman <action>",
	Short: "EC2 instance management tool",
	Long: `ec-manager (ecman) is a CLI tool that helps you manage your AWS EC2 instances.
It provides commands for:
- Creating new instances with proper configuration
- Listing and checking instance status
- Migrating instances to new AMIs
- Managing instance backups
- Cleaning up unused instances`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		// Find the command
		c, _, err := cmd.Root().Find(args)
		if err != nil {
			return fmt.Errorf("unknown command %q", args[0])
		}
		return c.Help()
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Set up logging
		logger.Init(logger.LogLevel(logLevel))

		// Set up mock mode
		client.SetMockMode(mockMode)
		if mockMode {
			client.SetMockClient(client.NewMockEC2Client())
		}

		return nil
	},
	SilenceUsage: true, // Don't show usage on errors
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var comps []string
		for _, c := range cmd.Root().Commands() {
			if !c.Hidden {
				comps = append(comps, c.Name())
			}
		}
		return comps, cobra.ShellCompDirectiveNoFileComp
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&mockMode, "mock", false, "Enable mock mode")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 5*time.Minute, "Timeout for AWS operations")
	rootCmd.PersistentFlags().BoolVar(&enabledFlag, "enabled", false, "Only process instances with ami-migrate=enabled tag")
	rootCmd.PersistentFlags().StringVar(&instanceID, "instance-id", "", "ID of the EC2 instance")
	rootCmd.PersistentFlags().StringVar(&newAMI, "new-ami", "", "ID of the new AMI to migrate to")
	rootCmd.PersistentFlags().StringVar(&userID, "user", "", "Your AWS username (defaults to current AWS user)")
}

// getUserID returns the user ID, either from flag or AWS credentials
func getUserID(cmd *cobra.Command) (string, error) {
	// Check if user flag is set
	user, err := cmd.Flags().GetString("user")
	if err != nil {
		return "", err
	}

	// If user flag is set, use it
	if user != "" {
		return user, nil
	}

	// If in mock mode, return a default user
	mock, _ := cmd.Flags().GetBool("mock")
	if mock {
		return "mock-user", nil
	}

	// Try to get user from AWS credentials
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	awsUser, err := config.GetAWSUsername(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get AWS username: %v", err)
	}

	if awsUser == "" {
		return "", fmt.Errorf("--user flag is required when AWS username cannot be determined")
	}

	return awsUser, nil
}

// GetLogLevel returns the log level from flags
func GetLogLevel() string {
	return logLevel
}

// GetTimeout returns the timeout from flags
func GetTimeout() time.Duration {
	return timeout
}
