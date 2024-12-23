package cmd

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/client"
)

// initAWSClients initializes AWS clients and returns an AMI service
func initAWSClients(ctx context.Context) (*ami.Service, error) {
	// Load AWS configuration
	cfg, err := client.LoadAWSConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(cfg)
	
	// Create and return AMI service
	return ami.NewService(ec2Client), nil
}
