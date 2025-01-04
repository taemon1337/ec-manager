package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start an EC2 instance",
	Long:  `Start a stopped EC2 instance by its instance ID.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		amiService := awsClient.GetAMIService()

		err := amiService.StartInstance(cmd.Context(), startInstanceID)
		if err != nil {
			return fmt.Errorf("failed to start instance: %w", err)
		}

		fmt.Printf("Successfully started instance %s\n", startInstanceID)
		return nil
	},
}

var (
	// Instance ID to start
	startInstanceID string
)

func init() {
	rootCmd.AddCommand(startCmd)

	// Add flags
	startCmd.Flags().StringVarP(&startInstanceID, "instance", "i", "", "Instance ID to start")

	// Mark required flags
	if err := startCmd.MarkFlagRequired("instance"); err != nil {
		// This should only happen during development if we make a mistake with flag names
		panic(fmt.Sprintf("failed to mark flag as required: %v", err))
	}
}
