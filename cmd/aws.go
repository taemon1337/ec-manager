package cmd

import (
	"context"
	"fmt"

	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/client"
)

// Common variables
var awsClient *client.Client

func init() {
	cfg := client.NewDefaultConfig()
	cfg.MockMode = mockMode

	// Only initialize AWS client if not in test mode
	if !mockMode {
		var err error
		awsClient, err = client.NewClient(cfg)
		if err != nil {
			fmt.Printf("Warning: Failed to initialize AWS client: %v\n", err)
		}
	}
}

// initAWSClients initializes AWS clients and returns an AMI service
func initAWSClients(ctx context.Context) (*ami.Service, error) {
	if mockMode {
		// Create a new mock client for testing
		cfg := client.NewDefaultConfig()
		cfg.MockMode = true
		var err error
		awsClient, err = client.NewClient(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create mock client: %w", err)
		}
	}

	if awsClient == nil {
		return nil, fmt.Errorf("AWS client not initialized")
	}

	// Create AMI service
	return ami.NewService(awsClient.GetEC2Client()), nil
}
