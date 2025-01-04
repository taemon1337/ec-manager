package mock

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/mock"
	"github.com/taemon1337/ec-manager/pkg/mock/fixtures"
	"github.com/taemon1337/ec-manager/pkg/mock/waiters"
)

var Anything = mock.Anything
var MatchedBy = mock.MatchedBy

// EC2ClientKey is the key used to store the EC2 client in context
type ec2ClientKey struct{}

var EC2ClientKey = ec2ClientKey{}

// STSClientKey is the key used to store the STS client in context
type stsClientKey struct{}

var STSClientKey = stsClientKey{}

// IAMClientKey is the key used to store the IAM client in context
type iamClientKey struct{}

var IAMClientKey = iamClientKey{}

// MockEC2Client is a mock implementation of the EC2Client interface
type MockEC2Client struct {
	mock.Mock
	InstanceStoppedWaiter *waiters.MockInstanceStoppedWaiter
	InstanceRunningWaiter *waiters.MockInstanceRunningWaiter
	VolumeAvailableWaiter *waiters.MockVolumeAvailableWaiter
}

// NewMockEC2Client creates a new mock EC2 client
func NewMockEC2Client(t *testing.T) *MockEC2Client {
	m := &MockEC2Client{}
	m.Test(t)
	return m
}

// NewMockEC2ClientWithoutT creates a new mock EC2 client without requiring a testing.T
func NewMockEC2ClientWithoutT() *MockEC2Client {
	return &MockEC2Client{}
}

// ExpectDescribeInstances sets up expectations for DescribeInstances
func (m *MockEC2Client) ExpectDescribeInstances(instanceIDs []string, instances []types.Instance) *mock.Call {
	input := &ec2.DescribeInstancesInput{}
	if len(instanceIDs) > 0 {
		input.InstanceIds = instanceIDs
	}

	return m.On("DescribeInstances", mock.Anything, input).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: instances,
			},
		},
	}, nil)
}

// ExpectDescribeImages sets up expectations for DescribeImages
func (m *MockEC2Client) ExpectDescribeImages(imageIDs []string, images []types.Image) *mock.Call {
	input := &ec2.DescribeImagesInput{}
	if len(imageIDs) > 0 {
		input.ImageIds = imageIDs
	}

	return m.On("DescribeImages", mock.Anything, input).Return(&ec2.DescribeImagesOutput{
		Images: images,
	}, nil)
}

// ExpectDescribeSubnets sets up expectations for DescribeSubnets
func (m *MockEC2Client) ExpectDescribeSubnets(subnetIDs []string, subnets []types.Subnet) *mock.Call {
	input := &ec2.DescribeSubnetsInput{}
	if len(subnetIDs) > 0 {
		input.SubnetIds = subnetIDs
	}

	return m.On("DescribeSubnets", mock.Anything, input).Return(&ec2.DescribeSubnetsOutput{
		Subnets: subnets,
	}, nil)
}

// DescribeImages implements the EC2 client interface
func (m *MockEC2Client) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(1) != nil {
		return nil, args.Get(1).(error)
	}

	if len(params.ImageIds) > 0 {
		// Specific AMI lookup
		return args.Get(0).(*ec2.DescribeImagesOutput), nil
	}

	// List operation
	return &ec2.DescribeImagesOutput{
		Images: fixtures.TestListAMIs(),
	}, nil
}

// DescribeInstances implements the EC2 client interface
func (m *MockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(1) != nil {
		return nil, args.Get(1).(error)
	}

	if len(params.InstanceIds) > 0 {
		// Specific instance lookup
		return args.Get(0).(*ec2.DescribeInstancesOutput), nil
	}

	// List operation
	instances := fixtures.TestListInstances()
	return &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: instances,
			},
		},
	}, nil
}

// DescribeSubnets implements the EC2 client interface
func (m *MockEC2Client) DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(1) != nil {
		return nil, args.Get(1).(error)
	}

	if len(params.SubnetIds) > 0 {
		// Specific subnet lookup
		return args.Get(0).(*ec2.DescribeSubnetsOutput), nil
	}

	// List operation
	return &ec2.DescribeSubnetsOutput{
		Subnets: fixtures.TestListSubnets(),
	}, nil
}

// DescribeKeyPairs implements the EC2 client interface
func (m *MockEC2Client) DescribeKeyPairs(ctx context.Context, params *ec2.DescribeKeyPairsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeKeyPairsOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(1) != nil {
		return nil, args.Get(1).(error)
	}

	if len(params.KeyNames) > 0 {
		// Specific key pair lookup
		return args.Get(0).(*ec2.DescribeKeyPairsOutput), nil
	}

	// List operation
	return &ec2.DescribeKeyPairsOutput{
		KeyPairs: fixtures.TestListKeyPairs(),
	}, nil
}

// CreateImage implements the EC2 client interface
func (m *MockEC2Client) CreateImage(ctx context.Context, params *ec2.CreateImageInput, optFns ...func(*ec2.Options)) (*ec2.CreateImageOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(1) != nil {
		return nil, args.Get(1).(error)
	}
	return args.Get(0).(*ec2.CreateImageOutput), nil
}

// CreateTags implements the EC2 client interface
func (m *MockEC2Client) CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	args := m.Called(ctx, params, mock.Anything)
	if args.Get(1) != nil {
		return nil, args.Get(1).(error)
	}
	return args.Get(0).(*ec2.CreateTagsOutput), nil
}

// RunInstances implements the EC2 client interface
func (m *MockEC2Client) RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(1) != nil {
		return nil, args.Get(1).(error)
	}

	// Mock instance creation
	return &ec2.RunInstancesOutput{
		Instances: []types.Instance{
			{
				InstanceId: aws.String("i-mock123"),
				State: &types.InstanceState{
					Name: types.InstanceStateNameRunning,
				},
				InstanceType: types.InstanceTypeT2Micro,
			},
		},
	}, nil
}

// TerminateInstances implements the EC2 client interface
func (m *MockEC2Client) TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(1) != nil {
		return nil, args.Get(1).(error)
	}
	return args.Get(0).(*ec2.TerminateInstancesOutput), nil
}

// AttachVolume implements the EC2 client interface
func (m *MockEC2Client) AttachVolume(ctx context.Context, params *ec2.AttachVolumeInput, optFns ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(1) != nil {
		return nil, args.Get(1).(error)
	}
	return args.Get(0).(*ec2.AttachVolumeOutput), nil
}

// CreateSnapshot implements the EC2 client interface
func (m *MockEC2Client) CreateSnapshot(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(1) != nil {
		return nil, args.Get(1).(error)
	}
	return args.Get(0).(*ec2.CreateSnapshotOutput), nil
}

// CreateVolume implements the EC2 client interface
func (m *MockEC2Client) CreateVolume(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(1) != nil {
		return nil, args.Get(1).(error)
	}
	return args.Get(0).(*ec2.CreateVolumeOutput), nil
}

// DescribeSnapshots implements the EC2 client interface
func (m *MockEC2Client) DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(1) != nil {
		return nil, args.Get(1).(error)
	}
	return args.Get(0).(*ec2.DescribeSnapshotsOutput), nil
}

// DescribeVolumes implements the EC2 client interface
func (m *MockEC2Client) DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(1) != nil {
		return nil, args.Get(1).(error)
	}
	return args.Get(0).(*ec2.DescribeVolumesOutput), nil
}

// StopInstances implements the EC2 client interface
func (m *MockEC2Client) StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(1) != nil {
		return nil, args.Get(1).(error)
	}
	return args.Get(0).(*ec2.StopInstancesOutput), nil
}

// StartInstances implements the EC2 client interface
func (m *MockEC2Client) StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(1) != nil {
		return nil, args.Get(1).(error)
	}
	return args.Get(0).(*ec2.StartInstancesOutput), nil
}

// NewInstanceRunningWaiter returns a mock running waiter
func (m *MockEC2Client) NewInstanceRunningWaiter() interface {
	Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceRunningWaiterOptions)) error
} {
	return m.InstanceRunningWaiter
}

// NewInstanceStoppedWaiter returns a mock stopped waiter
func (m *MockEC2Client) NewInstanceStoppedWaiter() interface {
	Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceStoppedWaiterOptions)) error
} {
	return m.InstanceStoppedWaiter
}

// NewInstanceTerminatedWaiter returns a mock terminated waiter
func (m *MockEC2Client) NewInstanceTerminatedWaiter() interface {
	Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceTerminatedWaiterOptions)) error
} {
	return &waiters.MockInstanceTerminatedWaiter{
		Mock: mock.Mock{},
	}
}

// NewVolumeAvailableWaiter returns a mock volume available waiter
func (m *MockEC2Client) NewVolumeAvailableWaiter() interface {
	Wait(ctx context.Context, params *ec2.DescribeVolumesInput, maxWaitDur time.Duration, optFns ...func(*ec2.VolumeAvailableWaiterOptions)) error
} {
	return m.VolumeAvailableWaiter
}

// MockSTSClient is a mock implementation of STSClient
type MockSTSClient struct {
	mock.Mock
}

// NewMockSTSClient creates a new mock STS client
func NewMockSTSClient(t *testing.T) *MockSTSClient {
	m := &MockSTSClient{}
	m.Test(t)
	return m
}

// NewMockSTSClientWithoutT creates a new mock STS client without testing.T
func NewMockSTSClientWithoutT() *MockSTSClient {
	return &MockSTSClient{}
}

// GetCallerIdentity implements the STS client interface
func (m *MockSTSClient) GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sts.GetCallerIdentityOutput), args.Error(1)
}

// AssumeRole implements the STS client interface
func (m *MockSTSClient) AssumeRole(ctx context.Context, params *sts.AssumeRoleInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sts.AssumeRoleOutput), args.Error(1)
}

// MockIAMClient is a mock implementation of IAMClient
type MockIAMClient struct {
	mock.Mock
}

// NewMockIAMClient creates a new mock IAM client
func NewMockIAMClient(t *testing.T) *MockIAMClient {
	m := &MockIAMClient{}
	m.Test(t)
	return m
}

// NewMockIAMClientWithoutT creates a new mock IAM client without testing.T
func NewMockIAMClientWithoutT() *MockIAMClient {
	return &MockIAMClient{}
}

// GetUser implements the IAM client interface
func (m *MockIAMClient) GetUser(ctx context.Context, params *iam.GetUserInput, optFns ...func(*iam.Options)) (*iam.GetUserOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*iam.GetUserOutput), args.Error(1)
}

// ListRoles implements the IAM client interface
func (m *MockIAMClient) ListRoles(ctx context.Context, params *iam.ListRolesInput, optFns ...func(*iam.Options)) (*iam.ListRolesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*iam.ListRolesOutput), args.Error(1)
}

// ListUsers implements the IAM client interface
func (m *MockIAMClient) ListUsers(ctx context.Context, params *iam.ListUsersInput, optFns ...func(*iam.Options)) (*iam.ListUsersOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*iam.ListUsersOutput), args.Error(1)
}
