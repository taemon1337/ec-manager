package cmd

import (
	"context"
	"fmt"

	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/client"
)

// Common variables
var awsClient *client.Client

// initAWSClients initializes AWS clients and returns an AMI service
func initAWSClients(ctx context.Context) (*ami.Service, error) {
	// Initialize client if not already done
	if awsClient == nil {
		cfg := client.NewDefaultConfig()
		cfg.MockMode = mockMode

		var err error
		awsClient, err = client.NewClient(cfg)
		if err != nil {
			if !mockMode {
				return nil, fmt.Errorf("failed to initialize AWS client: %w", err)
			}
			// In mock mode, we can proceed with a mock client even if AWS credentials are missing
		}
	}

	// Create AMI service
	return ami.NewService(awsClient.GetEC2Client()), nil
}
