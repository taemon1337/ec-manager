package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taemon1337/ec-manager/pkg/ami"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new EC2 instance",
	Long:  `Create a new EC2 instance from an AMI`,
	RunE:  runCreate,
}

var (
	amiID          string
	instanceType   string
	subnetID       string
	securityGroups []string
)

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&amiID, "ami", "a", "", "AMI ID to use")
	createCmd.Flags().StringVarP(&instanceType, "type", "t", "t2.micro", "Instance type")
	createCmd.Flags().StringVarP(&subnetID, "subnet", "s", "", "Subnet ID")
	createCmd.Flags().StringSliceVarP(&securityGroups, "security-groups", "g", nil, "Security group IDs")

	createCmd.MarkFlagRequired("ami")
}

func runCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Initialize AWS clients
	amiService, err := initAWSClients(ctx)
	if err != nil {
		return fmt.Errorf("init AWS clients: %w", err)
	}

	// Create instance config
	instanceConfig := ami.InstanceConfig{
		Name:   fmt.Sprintf("instance-%s", amiID),
		OSType: "linux", // We could detect this from the AMI
		Size:   instanceType,
	}

	// Create instance
	instance, err := amiService.CreateInstance(ctx, instanceConfig)
	if err != nil {
		return fmt.Errorf("create instance: %w", err)
	}

	fmt.Printf("Created instance %s\n", instance.InstanceID)
	return nil
}
