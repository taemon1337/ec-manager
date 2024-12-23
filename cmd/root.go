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
	instanceID string
	enabled    bool
	newAMI     string
	userID     string
	logLevel   string
	timeout    time.Duration
	defaultTimeout = 5 * time.Minute

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
		// Initialize logger
		logLevel := cmd.Flag("log-level").Value.String()
		logger.Init(logger.LogLevel(logLevel))

		// Set mock mode if enabled
		mock, err := cmd.Flags().GetBool("mock")
		if err != nil {
			return fmt.Errorf("failed to get mock flag: %w", err)
		}
		if mock {
			awsClient.SetMockMode(true)
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

// helpCmd represents the help command
var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "Show help for a command",
	Long: `Show detailed help and usage information for any command.
If no command is specified, shows help for all commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			rootCmd.Help()
			return
		}
		// Find the command
		c, _, err := rootCmd.Find(args)
		if err != nil {
			fmt.Printf("Unknown command %q\n", args[0])
			rootCmd.Help()
			return
		}
		c.Help()
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var comps []string
		for _, c := range rootCmd.Commands() {
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
	// Add help command
	rootCmd.AddCommand(helpCmd)

	// Add flags that are used by multiple commands
	rootCmd.PersistentFlags().StringVar(&instanceID, "instance-id", "", "ID of the EC2 instance")
	rootCmd.PersistentFlags().BoolVar(&enabled, "enabled", false, "Only process instances with ami-migrate=enabled tag")
	rootCmd.PersistentFlags().StringVar(&newAMI, "new-ami", "", "ID of the new AMI to migrate to")
	rootCmd.PersistentFlags().StringVar(&userID, "user", "", "Your AWS username (defaults to current AWS user)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", defaultTimeout, "Timeout for AWS operations")
	rootCmd.PersistentFlags().Bool("mock", false, "Enable mock mode")
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
