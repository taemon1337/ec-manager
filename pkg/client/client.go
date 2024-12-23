package client

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ecTypes "github.com/taemon1337/ec-manager/pkg/types"
)

var (
	defaultClient *Client
	mockMode      bool
	mockClient    ecTypes.EC2Client
)

// configLoaderInterface defines the interface for loading AWS config
type configLoaderInterface interface {
	LoadDefaultConfig(context.Context, ...func(*config.LoadOptions) error) (aws.Config, error)
}

// defaultConfigLoader implements configLoaderInterface using the actual AWS SDK
type defaultConfigLoader struct{}

func (d *defaultConfigLoader) LoadDefaultConfig(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx, optFns...)
}

// configLoader is used to load AWS config, can be replaced in tests
var configLoader configLoaderInterface = &defaultConfigLoader{}

// Client represents a client for interacting with AWS services
type Client struct {
	ec2Client ecTypes.EC2Client
}

func init() {
	defaultClient = NewClient()
}

// NewClient creates a new client
func NewClient() *Client {
	return &Client{}
}

// SetMockMode enables or disables mock mode
func SetMockMode(enabled bool) {
	mockMode = enabled
}

// SetMockClient sets the mock client
func SetMockClient(client ecTypes.EC2Client) {
	mockClient = client
}

// GetEC2Client returns an EC2 client
func GetEC2Client() ecTypes.EC2Client {
	if mockMode {
		if mockClient == nil {
			panic("mock client not set")
		}
		return mockClient
	}
	return nil
}

// GetEC2Client returns an EC2 client
func (c *Client) GetEC2Client(ctx context.Context) (ecTypes.EC2Client, error) {
	if mockMode {
		if mockClient == nil {
			return nil, fmt.Errorf("mock client not set")
		}
		return mockClient, nil
	}

	if c.ec2Client != nil {
		return c.ec2Client, nil
	}

	cfg, err := LoadAWSConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	c.ec2Client = ec2.NewFromConfig(cfg)
	return c.ec2Client, nil
}

// LoadAWSConfig loads the AWS configuration
func LoadAWSConfig(ctx context.Context) (aws.Config, error) {
	return configLoader.LoadDefaultConfig(ctx)
}

// LoadAWSConfig loads AWS configuration from environment variables or credentials file
func LoadAWSConfigWithRegion(ctx context.Context, region string) (aws.Config, error) {
	// Check if AWS credentials are set
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		return config.LoadDefaultConfig(ctx, config.WithRegion(region))
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return cfg, fmt.Errorf("unable to load AWS config: %w", err)
	}

	return cfg, nil
}

// NewMockEC2Client creates a new mock EC2 client with default outputs
func NewMockEC2Client() *ecTypes.MockEC2Client {
	return &ecTypes.MockEC2Client{
		DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId: aws.String("i-123"),
							State: &types.InstanceState{
								Name: types.InstanceStateNameRunning,
							},
						},
					},
				},
			},
		},
		StopInstancesOutput: &ec2.StopInstancesOutput{
			StoppingInstances: []types.InstanceStateChange{
				{
					CurrentState: &types.InstanceState{
						Name: types.InstanceStateNameStopped,
					},
					InstanceId: aws.String("i-123"),
				},
			},
		},
		StartInstancesOutput: &ec2.StartInstancesOutput{
			StartingInstances: []types.InstanceStateChange{
				{
					CurrentState: &types.InstanceState{
						Name: types.InstanceStateNameRunning,
					},
					InstanceId: aws.String("i-123"),
				},
			},
		},
		CreateImageOutput: &ec2.CreateImageOutput{
			ImageId: aws.String("ami-123"),
		},
		DescribeImagesOutput: &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId: aws.String("ami-123"),
					State:   types.ImageStateAvailable,
				},
			},
		},
		CreateTagsOutput: &ec2.CreateTagsOutput{},
		RunInstancesOutput: &ec2.RunInstancesOutput{
			Instances: []types.Instance{
				{
					InstanceId: aws.String("i-new"),
					State: &types.InstanceState{
						Name: types.InstanceStateNamePending,
					},
				},
			},
		},
		TerminateInstancesOutput: &ec2.TerminateInstancesOutput{
			TerminatingInstances: []types.InstanceStateChange{
				{
					CurrentState: &types.InstanceState{
						Name: types.InstanceStateNameShuttingDown,
					},
					InstanceId: aws.String("i-123"),
				},
			},
		},
		CreateSnapshotOutput: &ec2.CreateSnapshotOutput{
			SnapshotId: aws.String("snap-123"),
		},
		DescribeSnapshotsOutput: &ec2.DescribeSnapshotsOutput{
			Snapshots: []types.Snapshot{
				{
					SnapshotId: aws.String("snap-123"),
					State:      types.SnapshotStateCompleted,
				},
			},
		},
		CreateVolumeOutput: &ec2.CreateVolumeOutput{
			VolumeId: aws.String("vol-123"),
		},
		DescribeVolumesOutput: &ec2.DescribeVolumesOutput{
			Volumes: []types.Volume{
				{
					VolumeId: aws.String("vol-123"),
					State:    types.VolumeStateAvailable,
				},
			},
		},
	}
}
