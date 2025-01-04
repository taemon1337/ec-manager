package client

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/mock"
	ecTypes "github.com/taemon1337/ec-manager/pkg/types"
)

// Client represents a client for interacting with AWS services
type Client struct {
	cfg      aws.Config
	mockMode bool
	mockEC2  *mock.MockEC2Client
	realEC2  *ec2.Client
	profile  string
	region   string
}

// NewDefaultConfig returns a default configuration
func NewDefaultConfig() (bool, string, string) {
	return false, "", "us-east-1"
}

// NewClient creates a new AWS client
func NewClient(mockMode bool, profile, region string) (*Client, error) {
	client := &Client{
		mockMode: mockMode,
		profile:  profile,
		region:   region,
	}

	if mockMode {
		// Create mock EC2 client
		mockEC2 := mock.NewMockEC2ClientWithoutT()
		mock.SetupDefaultMockResponses(mockEC2)
		client.mockEC2 = mockEC2
		return client, nil
	}

	// Only load AWS config if not in mock mode
	cfg, err := client.loadAWSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	client.cfg = cfg

	// Create real EC2 client
	client.realEC2 = ec2.NewFromConfig(cfg)

	return client, nil
}

// loadAWSConfig loads the AWS configuration
func (c *Client) loadAWSConfig() (aws.Config, error) {
	var cfg aws.Config
	var err error

	opts := []func(*config.LoadOptions) error{
		config.WithRegion(c.region),
	}

	if c.profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(c.profile))
	}

	cfg, err = config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return cfg, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Verify credentials
	stsClient := sts.NewFromConfig(cfg)
	_, err = stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return cfg, fmt.Errorf("failed to verify AWS credentials: %w", err)
	}

	return cfg, nil
}

// GetEC2Client returns the EC2 client (either mock or real)
func (c *Client) GetEC2Client() ecTypes.EC2Client {
	if c.mockMode {
		return c.mockEC2
	}
	return &EC2ClientWrapper{c.realEC2}
}

// GetAMIService returns a new AMI service instance
func (c *Client) GetAMIService() *ami.Service {
	return ami.NewService(c.GetEC2Client())
}

// ListImages lists AMIs based on the provided filters
func (c *Client) ListImages(filters []types.Filter) ([]types.Image, error) {
	if c.mockMode {
		output, err := c.mockEC2.DescribeImages(context.Background(), &ec2.DescribeImagesInput{
			Filters: filters,
		})
		if err != nil {
			return nil, err
		}
		return output.Images, nil
	}

	input := &ec2.DescribeImagesInput{
		Filters: filters,
	}

	output, err := c.realEC2.DescribeImages(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	return output.Images, nil
}

// EC2Client is an interface for AWS EC2 client
type EC2Client interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
	CreateImage(ctx context.Context, params *ec2.CreateImageInput, optFns ...func(*ec2.Options)) (*ec2.CreateImageOutput, error)
	CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
	TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
	AttachVolume(ctx context.Context, params *ec2.AttachVolumeInput, optFns ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error)
	CreateSnapshot(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error)
	CreateVolume(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error)
	DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error)
	DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error)
	DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error)
	DescribeKeyPairs(ctx context.Context, params *ec2.DescribeKeyPairsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeKeyPairsOutput, error)
	StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
	StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error)
	NewInstanceRunningWaiter() interface {
		Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceRunningWaiterOptions)) error
	}
	NewInstanceStoppedWaiter() interface {
		Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceStoppedWaiterOptions)) error
	}
	NewInstanceTerminatedWaiter() interface {
		Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceTerminatedWaiterOptions)) error
	}
	NewVolumeAvailableWaiter() interface {
		Wait(ctx context.Context, params *ec2.DescribeVolumesInput, maxWaitDur time.Duration, optFns ...func(*ec2.VolumeAvailableWaiterOptions)) error
	}
}

// EC2ClientWrapper wraps the AWS SDK EC2 client to implement our EC2Client interface
type EC2ClientWrapper struct {
	*ec2.Client
}

// NewInstanceRunningWaiter implements EC2Client
func (c *EC2ClientWrapper) NewInstanceRunningWaiter() interface {
	Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceRunningWaiterOptions)) error
} {
	return ec2.NewInstanceRunningWaiter(c.Client)
}

// NewInstanceStoppedWaiter implements EC2Client
func (c *EC2ClientWrapper) NewInstanceStoppedWaiter() interface {
	Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceStoppedWaiterOptions)) error
} {
	return ec2.NewInstanceStoppedWaiter(c.Client)
}

// NewInstanceTerminatedWaiter implements EC2Client
func (c *EC2ClientWrapper) NewInstanceTerminatedWaiter() interface {
	Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceTerminatedWaiterOptions)) error
} {
	return ec2.NewInstanceTerminatedWaiter(c.Client)
}

// NewVolumeAvailableWaiter implements EC2Client
func (c *EC2ClientWrapper) NewVolumeAvailableWaiter() interface {
	Wait(ctx context.Context, params *ec2.DescribeVolumesInput, maxWaitDur time.Duration, optFns ...func(*ec2.VolumeAvailableWaiterOptions)) error
} {
	return ec2.NewVolumeAvailableWaiter(c.Client)
}
