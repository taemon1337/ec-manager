package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ami-migrate/pkg/logger"
)

var (
	// Common flags
	instanceID string
	enabled    bool
	newAMI     string
	logLevel   string
	timeout    time.Duration
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ami-migrate",
	Short: "A tool for migrating AWS EC2 instances to new AMIs",
	Long: `ami-migrate is a CLI tool that helps you migrate your AWS EC2 instances
to new AMIs. It can backup your instances, create new AMIs, and migrate
your instances to the new AMIs.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize logger
		switch logLevel {
		case "debug", "info", "warn", "error":
			logger.Init(logger.LogLevel(logLevel))
		default:
			return fmt.Errorf("invalid log level: %s", logLevel)
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Add common flags to root command
	rootCmd.PersistentFlags().StringVar(&instanceID, "instance-id", "", "ID of the EC2 instance")
	rootCmd.PersistentFlags().BoolVar(&enabled, "enabled", false, "Only process instances with ami-migrate=enabled tag")
	rootCmd.PersistentFlags().StringVar(&newAMI, "new-ami", "", "ID of the new AMI to migrate to")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 5*time.Minute, "Timeout for AWS operations (e.g., '5m', '1h')")
}

// GetLogLevel returns the log level from flags
func GetLogLevel() string {
	return logLevel
}

// GetTimeout returns the timeout from flags
func GetTimeout() time.Duration {
	return timeout
}
