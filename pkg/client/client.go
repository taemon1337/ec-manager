package client

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/taemon1337/ec-manager/pkg/mock"
	ecTypes "github.com/taemon1337/ec-manager/pkg/types"
)

// Client represents a client for interacting with AWS services
type Client struct {
	ec2Client ecTypes.EC2Client
	mockMode  bool
}

// Config holds the configuration for the client
type Config struct {
	Profile  string
	Region   string
	MockMode bool
}

// NewDefaultConfig returns a default configuration
func NewDefaultConfig() *Config {
	return &Config{
		Region: "us-east-1",
	}
}

// NewClient creates a new AWS client
func NewClient(cfg *Config) (*Client, error) {
	if cfg.MockMode {
		return &Client{
			ec2Client: mock.NewMockEC2Client(),
			mockMode:  true,
		}, nil
	}

	// Load AWS configuration
	awsCfg, err := loadAWSConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create EC2 client
	ec2Client := &EC2ClientWrapper{
		Client: ec2.NewFromConfig(awsCfg),
	}

	return &Client{
		ec2Client: ec2Client,
		mockMode:  cfg.MockMode,
	}, nil
}

// GetEC2Client returns the EC2 client
func (c *Client) GetEC2Client() ecTypes.EC2Client {
	return c.ec2Client
}

// ListImages lists AMIs based on the provided filters
func (c *Client) ListImages(filters []types.Filter) ([]types.Image, error) {
	if c.mockMode {
		return []types.Image{
			{
				ImageId:      aws.String("ami-123"),
				Name:        aws.String("test-ami"),
				Description: aws.String("Test AMI"),
				State:       types.ImageStateAvailable,
				Tags: []types.Tag{
					{
						Key:   aws.String("Project"),
						Value: aws.String("ec-manager"),
					},
				},
			},
		}, nil
	}

	input := &ec2.DescribeImagesInput{
		Filters: filters,
	}

	output, err := c.ec2Client.DescribeImages(context.Background(), input)
	if err != nil {
		return nil, err
	}

	return output.Images, nil
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

// DescribeSubnets implements EC2Client
func (c *EC2ClientWrapper) DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
	return c.Client.DescribeSubnets(ctx, params, optFns...)
}

// DescribeKeyPairs implements EC2Client
func (c *EC2ClientWrapper) DescribeKeyPairs(ctx context.Context, params *ec2.DescribeKeyPairsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeKeyPairsOutput, error) {
	return c.Client.DescribeKeyPairs(ctx, params, optFns...)
}

// loadAWSConfig loads the AWS configuration
func loadAWSConfig(cfg *Config) (aws.Config, error) {
	var optFns []func(*config.LoadOptions) error

	if cfg.Profile != "" {
		optFns = append(optFns, config.WithSharedConfigProfile(cfg.Profile))
	}

	if cfg.Region != "" {
		optFns = append(optFns, config.WithRegion(cfg.Region))
	}

	return config.LoadDefaultConfig(context.Background(), optFns...)
}

// AssumeRole assumes an IAM role
func AssumeRole(ctx context.Context, cfg aws.Config, roleARN string, mfaToken string) (*aws.Config, error) {
	stsSvc := sts.NewFromConfig(cfg)
	
	assumeRoleInput := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleARN),
		RoleSessionName: aws.String("ec-manager-session"),
	}

	if mfaToken != "" {
		assumeRoleInput.SerialNumber = aws.String("arn:aws:iam::ACCOUNT_ID:mfa/USER")
		assumeRoleInput.TokenCode = aws.String(mfaToken)
	}

	creds := stscreds.NewAssumeRoleProvider(stsSvc, roleARN)
	newCfg := cfg.Copy()
	newCfg.Credentials = aws.NewCredentialsCache(creds)

	return &newCfg, nil
}

// ListRoles lists available IAM roles
func ListRoles(configFile string) error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	log.Printf("Available roles in %s:\n%s", configFile, string(data))
	return nil
}
