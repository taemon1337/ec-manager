package mock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/mock"
	"github.com/taemon1337/ec-manager/pkg/mock/fixtures"
)

// Define a custom key type for context values
type ec2ClientContextKey struct{}

var EC2ClientContextKey = ec2ClientContextKey{}

const (
// Removed the ec2ClientKey constant as it's no longer used
)

// SetupDefaultMockResponses sets up common mock responses for testing
func SetupDefaultMockResponses(m *MockEC2Client) {
	// Setup default responses
	m.On("DescribeInstances", mock.Anything, mock.Anything).Return(&ec2.DescribeInstancesOutput{
		Reservations: []ec2types.Reservation{
			{
				Instances: fixtures.TestListInstances(),
			},
		},
	}, nil)

	m.On("DescribeImages", mock.Anything, mock.Anything).Return(&ec2.DescribeImagesOutput{
		Images: fixtures.TestListAMIs(),
	}, nil)

	m.On("DescribeSubnets", mock.Anything, mock.Anything).Return(&ec2.DescribeSubnetsOutput{
		Subnets: fixtures.TestListSubnets(),
	}, nil)

	m.On("DescribeKeyPairs", mock.Anything, mock.Anything).Return(&ec2.DescribeKeyPairsOutput{
		KeyPairs: fixtures.TestListKeyPairs(),
	}, nil)

	m.On("CreateImage", mock.Anything, mock.Anything).Return(&ec2.CreateImageOutput{}, nil)
	m.On("CreateTags", mock.Anything, mock.Anything).Return(&ec2.CreateTagsOutput{}, nil)
	m.On("RunInstances", mock.Anything, mock.Anything).Return(&ec2.RunInstancesOutput{}, nil)
	m.On("StopInstances", mock.Anything, mock.Anything).Return(&ec2.StopInstancesOutput{}, nil)
	m.On("StartInstances", mock.Anything, mock.Anything).Return(&ec2.StartInstancesOutput{}, nil)
	m.On("TerminateInstances", mock.Anything, mock.Anything).Return(&ec2.TerminateInstancesOutput{}, nil)
	m.On("AttachVolume", mock.Anything, mock.Anything).Return(&ec2.AttachVolumeOutput{}, nil)
	m.On("CreateSnapshot", mock.Anything, mock.Anything).Return(&ec2.CreateSnapshotOutput{}, nil)
	m.On("CreateVolume", mock.Anything, mock.Anything).Return(&ec2.CreateVolumeOutput{}, nil)
	m.On("DescribeSnapshots", mock.Anything, mock.Anything).Return(&ec2.DescribeSnapshotsOutput{}, nil)
	m.On("DescribeVolumes", mock.Anything, mock.Anything).Return(&ec2.DescribeVolumesOutput{}, nil)
}

// WithMockEC2Client creates a context with a mock EC2 client for testing
func WithMockEC2Client(ctx context.Context, setupFn func(*MockEC2Client)) context.Context {
	mockClient := NewMockEC2Client(nil)
	if setupFn != nil {
		setupFn(mockClient)
	}
	return context.WithValue(ctx, EC2ClientContextKey, mockClient)
}

// GetMockEC2Client retrieves the mock EC2 client from the context
func GetMockEC2Client(ctx context.Context) *MockEC2Client {
	if client, ok := ctx.Value(EC2ClientContextKey).(*MockEC2Client); ok {
		return client
	}
	return nil
}
