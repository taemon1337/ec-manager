package mock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// MockEC2Client is a mock implementation of EC2Client
type MockEC2Client struct {
	// Outputs
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

	// Functions
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

// NewMockEC2Client creates a new mock EC2 client
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

// DescribeSubnets implements EC2Client
func (m *MockEC2Client) DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
	if m.DescribeSubnetsFunc != nil {
		return m.DescribeSubnetsFunc(ctx, params, optFns...)
	}
	return m.DescribeSubnetsOutput, nil
}

// DescribeKeyPairs implements EC2Client
func (m *MockEC2Client) DescribeKeyPairs(ctx context.Context, params *ec2.DescribeKeyPairsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeKeyPairsOutput, error) {
	if m.DescribeKeyPairsFunc != nil {
		return m.DescribeKeyPairsFunc(ctx, params, optFns...)
	}
	return m.DescribeKeyPairsOutput, nil
}
