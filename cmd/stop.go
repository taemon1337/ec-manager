package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var stopInstanceID string

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop an EC2 instance",
	Long:  `Stop a running EC2 instance by its instance ID.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if stopInstanceID == "" {
			return fmt.Errorf("instance ID must be set")
		}

		ctx := cmd.Context()
		amiService := awsClient.GetAMIService()

		fmt.Printf("Stopping instance %s...\n", stopInstanceID)
		err := amiService.StopInstance(ctx, stopInstanceID)
		if err != nil {
			return fmt.Errorf("failed to stop instance: %v", err)
		}

		fmt.Printf("Successfully stopped instance %s\n", stopInstanceID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.Flags().StringVarP(&stopInstanceID, "instance", "i", "", "Instance ID to stop")
	if err := stopCmd.MarkFlagRequired("instance"); err != nil {
		// This should only happen during development if we make a mistake with flag names
		panic(fmt.Sprintf("failed to mark flag as required: %v", err))
	}
}
