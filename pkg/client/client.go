package client

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/taemon1337/ec-manager/pkg/types"
)

var defaultClient *Client

type Client struct {
	cfg        aws.Config
	ready      bool
	mockMode   bool
	mockClient types.EC2Client
}

func init() {
	defaultClient = NewClient()
}

// NewClient creates a new AWS client
func NewClient() *Client {
	return &Client{}
}

// SetMockMode enables or disables mock mode
func SetMockMode(enabled bool) {
	defaultClient.mockMode = enabled
}

// SetMockEC2Client sets the mock EC2 client for testing
func SetMockEC2Client(client types.EC2Client) {
	defaultClient.mockClient = client
}

// SetMockMode sets the mock mode for the client
func (c *Client) SetMockMode(enabled bool) {
	c.mockMode = enabled
}

// SetEC2Client sets the EC2 client for the client
func (c *Client) SetEC2Client(client types.EC2Client) {
	c.mockClient = client
}

// GetEC2Client returns an EC2 client
func (c *Client) GetEC2Client(ctx context.Context) (types.EC2Client, error) {
	if c.mockMode {
		if c.mockClient == nil {
			return nil, fmt.Errorf("mock client not set")
		}
		return c.mockClient, nil
	}

	if !c.ready {
		cfg, err := LoadAWSConfig(ctx)
		if err != nil {
			return nil, err
		}
		c.cfg = cfg
		c.ready = true
	}

	return ec2.NewFromConfig(c.cfg), nil
}

// LoadAWSConfig loads AWS configuration from environment variables or credentials file
func LoadAWSConfig(ctx context.Context) (aws.Config, error) {
	// Check if AWS credentials are set
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		return config.LoadDefaultConfig(ctx, config.WithRegion("us-west-2"))
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return cfg, fmt.Errorf("unable to load AWS config: %w", err)
	}

	return cfg, nil
}
