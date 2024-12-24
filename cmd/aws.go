package cmd

import (
	"context"
	"fmt"

	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/client"
)

// initAWSClients initializes AWS clients and returns an AMI service
func initAWSClients(ctx context.Context) (*ami.Service, error) {
	// Get EC2 client
	c := client.NewClient()
	ec2Client, err := c.GetEC2Client(ctx)
	if err != nil {
		return nil, fmt.Errorf("get EC2 client: %w", err)
	}

	// Create AMI service
	return ami.NewService(ec2Client), nil
}
