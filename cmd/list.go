package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List EC2 instances",
	Long:  `List all EC2 instances and their current state`,
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Initialize AWS clients
	amiService, err := initAWSClients(ctx)
	if err != nil {
		return fmt.Errorf("init AWS clients: %w", err)
	}

	// List instances
	instances, err := amiService.ListUserInstances(ctx, "")
	if err != nil {
		return fmt.Errorf("list instances: %w", err)
	}

	// Print instances
	for _, instance := range instances {
		fmt.Printf("Instance %s status: %s\n", instance.InstanceID, instance.State)
	}

	return nil
}
