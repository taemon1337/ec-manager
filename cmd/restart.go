package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var restartInstanceID string

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart an EC2 instance",
	Long:  `Restart an EC2 instance by stopping and starting it.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if restartInstanceID == "" {
			return fmt.Errorf("instance ID must be set")
		}

		ctx := cmd.Context()
		amiService := awsClient.GetAMIService()

		fmt.Printf("Restarting instance %s...\n", restartInstanceID)
		err := amiService.RestartInstance(ctx, restartInstanceID)
		if err != nil {
			return fmt.Errorf("failed to restart instance: %v", err)
		}

		fmt.Printf("Successfully restarted instance %s\n", restartInstanceID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
	restartCmd.Flags().StringVarP(&restartInstanceID, "instance", "i", "", "Instance ID to restart")
	if err := restartCmd.MarkFlagRequired("instance"); err != nil {
		// This should only happen during development if we make a mistake with flag names
		panic(fmt.Sprintf("failed to mark flag as required: %v", err))
	}
}
