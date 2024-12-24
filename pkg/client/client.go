package client

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/taemon1337/ec-manager/pkg/mock"
	ecTypes "github.com/taemon1337/ec-manager/pkg/types"
)

// Client represents a client for interacting with AWS services
type Client struct {
	EC2Client ecTypes.EC2Client
}

// Config holds the configuration for the client
type Config struct {
	Region         string
	Profile        string
	Timeout        time.Duration
	RetryCount     int
	MockMode       bool
	InstanceConfig *InstanceConfig
}

// InstanceConfig holds the configuration for EC2 instances
type InstanceConfig struct {
	InstanceType     string
	SubnetID         string
	SecurityGroupIDs []string
}

// NewDefaultConfig returns a default configuration
func NewDefaultConfig() *Config {
	return &Config{
		Region:     "us-west-2",
		Profile:    "default",
		Timeout:    5 * time.Minute,
		RetryCount: 3,
		InstanceConfig: &InstanceConfig{
			InstanceType: "t2.micro",
		},
	}
}

// NewClient creates a new client
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = NewDefaultConfig()
	}

	var ec2Client ecTypes.EC2Client

	if cfg.MockMode {
		// In mock mode, always create a mock client
		ec2Client = mock.NewMockEC2Client()
		return &Client{
			EC2Client: ec2Client,
		}, nil
	}

	// Only try to load AWS config if not in mock mode
	awsCfg, err := loadAWSConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	ec2Client = &EC2ClientWrapper{
		Client: ec2.NewFromConfig(awsCfg),
	}

	return &Client{
		EC2Client: ec2Client,
	}, nil
}

// loadAWSConfig loads the AWS configuration
func loadAWSConfig(cfg *Config) (aws.Config, error) {
	optFns := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
	}

	if cfg.Profile != "" {
		optFns = append(optFns, config.WithSharedConfigProfile(cfg.Profile))
	}

	return config.LoadDefaultConfig(context.TODO(), optFns...)
}

// GetEC2Client returns the EC2 client
func (c *Client) GetEC2Client() ecTypes.EC2Client {
	return c.EC2Client
}

// GetInstanceConfig returns the instance configuration
func (c *Config) GetInstanceConfig() *InstanceConfig {
	if c.InstanceConfig == nil {
		return &InstanceConfig{
			InstanceType: "t2.micro",
		}
	}
	return c.InstanceConfig
}

// GetTimeout returns the timeout duration
func (c *Config) GetTimeout() time.Duration {
	if c.Timeout == 0 {
		return 5 * time.Minute
	}
	return c.Timeout
}

// EC2ClientWrapper wraps the AWS SDK EC2 client to implement our EC2Client interface
type EC2ClientWrapper struct {
	*ec2.Client
}

// NewInstanceRunningWaiter implements EC2Client
func (c *EC2ClientWrapper) NewInstanceRunningWaiter() *ec2.InstanceRunningWaiter {
	return ec2.NewInstanceRunningWaiter(c.Client)
}

// NewInstanceStoppedWaiter implements EC2Client
func (c *EC2ClientWrapper) NewInstanceStoppedWaiter() *ec2.InstanceStoppedWaiter {
	return ec2.NewInstanceStoppedWaiter(c.Client)
}

// NewInstanceTerminatedWaiter implements EC2Client
func (c *EC2ClientWrapper) NewInstanceTerminatedWaiter() *ec2.InstanceTerminatedWaiter {
	return ec2.NewInstanceTerminatedWaiter(c.Client)
}

// NewVolumeAvailableWaiter implements EC2Client
func (c *EC2ClientWrapper) NewVolumeAvailableWaiter() *ec2.VolumeAvailableWaiter {
	return ec2.NewVolumeAvailableWaiter(c.Client)
}
