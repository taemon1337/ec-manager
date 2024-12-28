package mock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/mock"
)

var Anything = mock.Anything
var MatchedBy = mock.MatchedBy

// EC2ClientKey is the key used to store the EC2 client in context
type ec2ClientKey struct{}
var EC2ClientKey = ec2ClientKey{}

type MockEC2Client struct {
	mock.Mock

	// Outputs (for backward compatibility)
	DescribeInstancesOutput  *ec2.DescribeInstancesOutput
	StopInstancesOutput      *ec2.StopInstancesOutput
	StartInstancesOutput     *ec2.StartInstancesOutput
	CreateImageOutput        *ec2.CreateImageOutput
	DescribeImagesOutput     *ec2.DescribeImagesOutput
	CreateTagsOutput         *ec2.CreateTagsOutput
	RunInstancesOutput       *ec2.RunInstancesOutput
	TerminateInstancesOutput *ec2.TerminateInstancesOutput
	AttachVolumeOutput       *ec2.AttachVolumeOutput
	CreateSnapshotOutput     *ec2.CreateSnapshotOutput
	CreateVolumeOutput       *ec2.CreateVolumeOutput
	DescribeSnapshotsOutput  *ec2.DescribeSnapshotsOutput
	DescribeVolumesOutput    *ec2.DescribeVolumesOutput
	DescribeSubnetsOutput    *ec2.DescribeSubnetsOutput
	DescribeKeyPairsOutput   *ec2.DescribeKeyPairsOutput

	// Functions (for backward compatibility)
	DescribeInstancesFunc    func(context.Context, *ec2.DescribeInstancesInput, ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	StopInstancesFunc        func(context.Context, *ec2.StopInstancesInput, ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
	StartInstancesFunc       func(context.Context, *ec2.StartInstancesInput, ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error)
	CreateImageFunc          func(context.Context, *ec2.CreateImageInput, ...func(*ec2.Options)) (*ec2.CreateImageOutput, error)
	DescribeImagesFunc       func(context.Context, *ec2.DescribeImagesInput, ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
	CreateTagsFunc           func(context.Context, *ec2.CreateTagsInput, ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	RunInstancesFunc         func(context.Context, *ec2.RunInstancesInput, ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
	TerminateInstancesFunc   func(context.Context, *ec2.TerminateInstancesInput, ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
	AttachVolumeFunc         func(context.Context, *ec2.AttachVolumeInput, ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error)
	CreateSnapshotFunc       func(context.Context, *ec2.CreateSnapshotInput, ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error)
	CreateVolumeFunc         func(context.Context, *ec2.CreateVolumeInput, ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error)
	DescribeSnapshotsFunc    func(context.Context, *ec2.DescribeSnapshotsInput, ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error)
	DescribeVolumesFunc      func(context.Context, *ec2.DescribeVolumesInput, ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error)
	DescribeSubnetsFunc      func(context.Context, *ec2.DescribeSubnetsInput, ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error)
	DescribeKeyPairsFunc     func(context.Context, *ec2.DescribeKeyPairsInput, ...func(*ec2.Options)) (*ec2.DescribeKeyPairsOutput, error)
}

func NewMockEC2Client() *MockEC2Client {
	return &MockEC2Client{
		DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId:   aws.String("i-123"),
							InstanceType: types.InstanceTypeT2Micro,
							KeyName:      aws.String("test-key"),
							SubnetId:     aws.String("subnet-123"),
							State: &types.InstanceState{
								Name: types.InstanceStateNameRunning,
							},
							Placement: &types.Placement{
								AvailabilityZone: aws.String("us-west-2a"),
							},
							Tags: []types.Tag{
								{
									Key:   aws.String("ami-migrate"),
									Value: aws.String("enabled"),
								},
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
					ImageId:      aws.String("ami-123"),
					Name:        aws.String("test-ami"),
					Description: aws.String("Test AMI"),
					State:      types.ImageStateAvailable,
					CreationDate: aws.String("2024-12-24T00:00:00Z"),
					Tags: []types.Tag{
						{
							Key:   aws.String("ami-migrate"),
							Value: aws.String("latest"),
						},
						{
							Key:   aws.String("Project"),
							Value: aws.String("ec-manager"),
						},
					},
				},
			},
		},
		CreateTagsOutput: &ec2.CreateTagsOutput{},
		RunInstancesOutput: &ec2.RunInstancesOutput{
			Instances: []types.Instance{
				{
					InstanceId: aws.String("i-123"),
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
		AttachVolumeOutput: &ec2.AttachVolumeOutput{
			Device:     aws.String("/dev/sda1"),
			InstanceId: aws.String("i-123"),
			VolumeId:   aws.String("vol-123"),
			State:      types.VolumeAttachmentStateAttached,
		},
		CreateSnapshotOutput: &ec2.CreateSnapshotOutput{
			SnapshotId: aws.String("snap-123"),
		},
		CreateVolumeOutput: &ec2.CreateVolumeOutput{
			VolumeId: aws.String("vol-123"),
		},
		DescribeSnapshotsOutput: &ec2.DescribeSnapshotsOutput{
			Snapshots: []types.Snapshot{
				{
					SnapshotId: aws.String("snap-123"),
					State:      types.SnapshotStateCompleted,
				},
			},
		},
		DescribeVolumesOutput: &ec2.DescribeVolumesOutput{
			Volumes: []types.Volume{
				{
					VolumeId: aws.String("vol-123"),
					State:    types.VolumeStateAvailable,
				},
			},
		},
		DescribeSubnetsOutput: &ec2.DescribeSubnetsOutput{
			Subnets: []types.Subnet{
				{
					SubnetId:     aws.String("subnet-123"),
					VpcId:        aws.String("vpc-123"),
					CidrBlock:    aws.String("172.31.0.0/20"),
					AvailabilityZone: aws.String("us-west-2a"),
					AvailableIpAddressCount: aws.Int32(4091),
					Tags: []types.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String("Main Subnet"),
						},
					},
				},
			},
		},
		DescribeKeyPairsOutput: &ec2.DescribeKeyPairsOutput{
			KeyPairs: []types.KeyPairInfo{
				{
					KeyName:        aws.String("test-key"),
					KeyFingerprint: aws.String("12:34:56:78:90:ab:cd:ef"),
					Tags: []types.Tag{
						{
							Key:   aws.String("Project"),
							Value: aws.String("ec-manager"),
						},
					},
				},
			},
		},
	}
}

// ExpectDescribeInstances sets up expectations for DescribeInstances
func (m *MockEC2Client) ExpectDescribeInstances(instanceID string, instance types.Instance, err error) {
	m.On("DescribeInstances", mock.Anything, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{instance},
			},
		},
	}, err)
}

// ExpectDescribeImages sets up expectations for DescribeImages
func (m *MockEC2Client) ExpectDescribeImages(imageID string, image types.Image, err error) {
	m.On("DescribeImages", mock.Anything, &ec2.DescribeImagesInput{
		ImageIds: []string{imageID},
	}).Return(&ec2.DescribeImagesOutput{
		Images: []types.Image{image},
	}, err)
}

// ExpectCreateImage sets up expectations for CreateImage
func (m *MockEC2Client) ExpectCreateImage(instanceID string, name string, imageID string, err error) {
	m.On("CreateImage", mock.Anything, &ec2.CreateImageInput{
		InstanceId:  aws.String(instanceID),
		Name:        aws.String(name),
		Description: aws.String("Backup of instance " + instanceID),
	}).Return(&ec2.CreateImageOutput{
		ImageId: aws.String(imageID),
	}, err)
}

// ExpectCreateTags sets up expectations for CreateTags
func (m *MockEC2Client) ExpectCreateTags(resourceID string, tags []types.Tag, err error) {
	m.On("CreateTags", mock.Anything, &ec2.CreateTagsInput{
		Resources: []string{resourceID},
		Tags:      tags,
	}).Return(&ec2.CreateTagsOutput{}, err)
}

// ExpectRunInstances sets up expectations for RunInstances
func (m *MockEC2Client) ExpectRunInstances(imageID string, instanceType types.InstanceType, instanceID string, err error) {
	m.On("RunInstances", mock.Anything, &ec2.RunInstancesInput{
		ImageId:      aws.String(imageID),
		InstanceType: instanceType,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	}).Return(&ec2.RunInstancesOutput{
		Instances: []types.Instance{
			{
				InstanceId: aws.String(instanceID),
			},
		},
	}, err)
}

// DescribeInstances implements the EC2 client interface
func (m *MockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*ec2.DescribeInstancesOutput), args.Error(1)
	}
	
	if m.DescribeInstancesFunc != nil {
		return m.DescribeInstancesFunc(ctx, params, optFns...)
	}
	
	return m.DescribeInstancesOutput, nil
}

// DescribeImages implements the EC2 client interface
func (m *MockEC2Client) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*ec2.DescribeImagesOutput), args.Error(1)
	}
	
	if m.DescribeImagesFunc != nil {
		return m.DescribeImagesFunc(ctx, params, optFns...)
	}
	
	return m.DescribeImagesOutput, nil
}

// CreateImage implements the EC2 client interface
func (m *MockEC2Client) CreateImage(ctx context.Context, params *ec2.CreateImageInput, optFns ...func(*ec2.Options)) (*ec2.CreateImageOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*ec2.CreateImageOutput), args.Error(1)
	}
	
	if m.CreateImageFunc != nil {
		return m.CreateImageFunc(ctx, params, optFns...)
	}
	
	return m.CreateImageOutput, nil
}

// CreateTags implements the EC2 client interface
func (m *MockEC2Client) CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*ec2.CreateTagsOutput), args.Error(1)
	}
	
	if m.CreateTagsFunc != nil {
		return m.CreateTagsFunc(ctx, params, optFns...)
	}
	
	return m.CreateTagsOutput, nil
}

// RunInstances implements the EC2 client interface
func (m *MockEC2Client) RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*ec2.RunInstancesOutput), args.Error(1)
	}
	
	if m.RunInstancesFunc != nil {
		return m.RunInstancesFunc(ctx, params, optFns...)
	}
	
	return m.RunInstancesOutput, nil
}

// TerminateInstances implements the EC2 client interface
func (m *MockEC2Client) TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*ec2.TerminateInstancesOutput), args.Error(1)
	}
	
	if m.TerminateInstancesFunc != nil {
		return m.TerminateInstancesFunc(ctx, params, optFns...)
	}
	
	return m.TerminateInstancesOutput, nil
}

// AttachVolume implements the EC2 client interface
func (m *MockEC2Client) AttachVolume(ctx context.Context, params *ec2.AttachVolumeInput, optFns ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*ec2.AttachVolumeOutput), args.Error(1)
	}
	
	if m.AttachVolumeFunc != nil {
		return m.AttachVolumeFunc(ctx, params, optFns...)
	}
	
	return m.AttachVolumeOutput, nil
}

// CreateSnapshot implements the EC2 client interface
func (m *MockEC2Client) CreateSnapshot(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*ec2.CreateSnapshotOutput), args.Error(1)
	}
	
	if m.CreateSnapshotFunc != nil {
		return m.CreateSnapshotFunc(ctx, params, optFns...)
	}
	
	return m.CreateSnapshotOutput, nil
}

// CreateVolume implements the EC2 client interface
func (m *MockEC2Client) CreateVolume(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*ec2.CreateVolumeOutput), args.Error(1)
	}
	
	if m.CreateVolumeFunc != nil {
		return m.CreateVolumeFunc(ctx, params, optFns...)
	}
	
	return m.CreateVolumeOutput, nil
}

// DescribeSnapshots implements the EC2 client interface
func (m *MockEC2Client) DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*ec2.DescribeSnapshotsOutput), args.Error(1)
	}
	
	if m.DescribeSnapshotsFunc != nil {
		return m.DescribeSnapshotsFunc(ctx, params, optFns...)
	}
	
	return m.DescribeSnapshotsOutput, nil
}

// DescribeVolumes implements the EC2 client interface
func (m *MockEC2Client) DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*ec2.DescribeVolumesOutput), args.Error(1)
	}
	
	if m.DescribeVolumesFunc != nil {
		return m.DescribeVolumesFunc(ctx, params, optFns...)
	}
	
	return m.DescribeVolumesOutput, nil
}

// DescribeSubnets implements the EC2 client interface
func (m *MockEC2Client) DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*ec2.DescribeSubnetsOutput), args.Error(1)
	}
	
	if m.DescribeSubnetsFunc != nil {
		return m.DescribeSubnetsFunc(ctx, params, optFns...)
	}
	
	return m.DescribeSubnetsOutput, nil
}

// DescribeKeyPairs implements the EC2 client interface
func (m *MockEC2Client) DescribeKeyPairs(ctx context.Context, params *ec2.DescribeKeyPairsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeKeyPairsOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*ec2.DescribeKeyPairsOutput), args.Error(1)
	}
	
	if m.DescribeKeyPairsFunc != nil {
		return m.DescribeKeyPairsFunc(ctx, params, optFns...)
	}
	
	return m.DescribeKeyPairsOutput, nil
}

// StopInstances implements the EC2 client interface
func (m *MockEC2Client) StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*ec2.StopInstancesOutput), args.Error(1)
	}
	
	if m.StopInstancesFunc != nil {
		return m.StopInstancesFunc(ctx, params, optFns...)
	}
	
	return m.StopInstancesOutput, nil
}

// StartInstances implements the EC2 client interface
func (m *MockEC2Client) StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(*ec2.StartInstancesOutput), args.Error(1)
	}
	
	if m.StartInstancesFunc != nil {
		return m.StartInstancesFunc(ctx, params, optFns...)
	}
	
	return m.StartInstancesOutput, nil
}
