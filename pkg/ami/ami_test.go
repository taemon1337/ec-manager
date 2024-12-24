package ami

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
)

type mockEC2Client struct {
	DescribeInstancesFunc  func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	DescribeImagesFunc     func(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
	CreateTagsFunc         func(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	RunInstancesFunc       func(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
	StopInstancesFunc      func(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
	StartInstancesFunc     func(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error)
	AttachVolumeFunc       func(ctx context.Context, params *ec2.AttachVolumeInput, optFns ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error)
	CreateSnapshotFunc     func(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error)
	TerminateInstancesFunc func(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
	CreateVolumeFunc       func(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error)
	CreateImageFunc        func(ctx context.Context, params *ec2.CreateImageInput, optFns ...func(*ec2.Options)) (*ec2.CreateImageOutput, error)
	DescribeSnapshotsFunc  func(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error)
	DescribeVolumesFunc    func(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error)
	DescribeSubnetsFunc    func(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error)
	DescribeKeyPairsFunc   func(ctx context.Context, params *ec2.DescribeKeyPairsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeKeyPairsOutput, error)
}

func (m *mockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if m.DescribeInstancesFunc != nil {
		return m.DescribeInstancesFunc(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockEC2Client) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	if m.DescribeImagesFunc != nil {
		return m.DescribeImagesFunc(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockEC2Client) CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	if m.CreateTagsFunc != nil {
		return m.CreateTagsFunc(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockEC2Client) RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	if m.RunInstancesFunc != nil {
		return m.RunInstancesFunc(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockEC2Client) StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
	if m.StopInstancesFunc != nil {
		return m.StopInstancesFunc(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockEC2Client) StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error) {
	if m.StartInstancesFunc != nil {
		return m.StartInstancesFunc(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockEC2Client) AttachVolume(ctx context.Context, params *ec2.AttachVolumeInput, optFns ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error) {
	if m.AttachVolumeFunc != nil {
		return m.AttachVolumeFunc(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockEC2Client) CreateSnapshot(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error) {
	if m.CreateSnapshotFunc != nil {
		return m.CreateSnapshotFunc(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockEC2Client) TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	if m.TerminateInstancesFunc != nil {
		return m.TerminateInstancesFunc(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockEC2Client) CreateVolume(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error) {
	if m.CreateVolumeFunc != nil {
		return m.CreateVolumeFunc(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockEC2Client) CreateImage(ctx context.Context, params *ec2.CreateImageInput, optFns ...func(*ec2.Options)) (*ec2.CreateImageOutput, error) {
	if m.CreateImageFunc != nil {
		return m.CreateImageFunc(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockEC2Client) DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error) {
	if m.DescribeSnapshotsFunc != nil {
		return m.DescribeSnapshotsFunc(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockEC2Client) DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error) {
	if m.DescribeVolumesFunc != nil {
		return m.DescribeVolumesFunc(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockEC2Client) DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
	if m.DescribeSubnetsFunc != nil {
		return m.DescribeSubnetsFunc(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockEC2Client) DescribeKeyPairs(ctx context.Context, params *ec2.DescribeKeyPairsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeKeyPairsOutput, error) {
	if m.DescribeKeyPairsFunc != nil {
		return m.DescribeKeyPairsFunc(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockEC2Client) NewInstanceRunningWaiter() *ec2.InstanceRunningWaiter {
	return nil
}

func (m *mockEC2Client) NewInstanceStoppedWaiter() *ec2.InstanceStoppedWaiter {
	return nil
}

func (m *mockEC2Client) NewInstanceTerminatedWaiter() *ec2.InstanceTerminatedWaiter {
	return nil
}

func (m *mockEC2Client) NewVolumeAvailableWaiter() *ec2.VolumeAvailableWaiter {
	return nil
}

func TestBackupInstance(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := &mockEC2Client{}
		svc := NewService(mockClient)

		mockClient.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
			return &ec2.DescribeInstancesOutput{
				Reservations: []types.Reservation{
					{
						Instances: []types.Instance{
							{
								InstanceId: aws.String("i-1234567890abcdef0"),
								State: &types.InstanceState{
									Name: types.InstanceStateNameRunning,
								},
							},
						},
					},
				},
			}, nil
		}

		mockClient.CreateImageFunc = func(ctx context.Context, params *ec2.CreateImageInput, optFns ...func(*ec2.Options)) (*ec2.CreateImageOutput, error) {
			return &ec2.CreateImageOutput{
				ImageId: aws.String("ami-1234567890abcdef0"),
			}, nil
		}

		mockClient.CreateTagsFunc = func(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
			return &ec2.CreateTagsOutput{}, nil
		}

		ctx := context.Background()
		imageID, err := svc.BackupInstance(ctx, "i-1234567890abcdef0")
		assert.NoError(t, err)
		assert.Equal(t, "ami-1234567890abcdef0", imageID)
	})

	t.Run("error", func(t *testing.T) {
		mockClient := &mockEC2Client{}
		svc := NewService(mockClient)

		mockClient.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
			return nil, fmt.Errorf("API error")
		}

		ctx := context.Background()
		imageID, err := svc.BackupInstance(ctx, "i-1234567890abcdef0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
		assert.Empty(t, imageID)
	})
}

func TestMigrateInstance(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := &mockEC2Client{}
		svc := NewService(mockClient)

		mockClient.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
			return &ec2.DescribeInstancesOutput{
				Reservations: []types.Reservation{
					{
						Instances: []types.Instance{
							{
								InstanceId:   aws.String("i-1234567890abcdef0"),
								InstanceType: types.InstanceTypeT2Micro,
								KeyName:      aws.String("test-key"),
								SubnetId:     aws.String("subnet-123"),
								State: &types.InstanceState{
									Name: types.InstanceStateNameRunning,
								},
							},
						},
					},
				},
			}, nil
		}

		mockClient.RunInstancesFunc = func(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
			return &ec2.RunInstancesOutput{
				Instances: []types.Instance{
					{
						InstanceId: aws.String("i-0987654321fedcba0"),
					},
				},
			}, nil
		}

		mockClient.CreateTagsFunc = func(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
			return &ec2.CreateTagsOutput{}, nil
		}

		ctx := context.Background()
		newInstanceID, err := svc.MigrateInstance(ctx, "i-1234567890abcdef0", "ami-1234567890abcdef0")
		assert.NoError(t, err)
		assert.Equal(t, "i-0987654321fedcba0", newInstanceID)
	})

	t.Run("error", func(t *testing.T) {
		mockClient := &mockEC2Client{}
		svc := NewService(mockClient)

		mockClient.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
			return nil, fmt.Errorf("API error")
		}

		ctx := context.Background()
		newInstanceID, err := svc.MigrateInstance(ctx, "i-1234567890abcdef0", "ami-1234567890abcdef0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
		assert.Empty(t, newInstanceID)
	})
}

func TestDeleteInstance(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := &mockEC2Client{}
		svc := NewService(mockClient)

		mockClient.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
			return &ec2.DescribeInstancesOutput{
				Reservations: []types.Reservation{
					{
						Instances: []types.Instance{
							{
								InstanceId: aws.String("i-1234567890abcdef0"),
								State: &types.InstanceState{
									Name: types.InstanceStateNameRunning,
								},
							},
						},
					},
				},
			}, nil
		}

		mockClient.TerminateInstancesFunc = func(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
			return &ec2.TerminateInstancesOutput{
				TerminatingInstances: []types.InstanceStateChange{
					{
						CurrentState: &types.InstanceState{
							Name: types.InstanceStateNameShuttingDown,
						},
						InstanceId: aws.String("i-1234567890abcdef0"),
					},
				},
			}, nil
		}

		ctx := context.Background()
		state, err := svc.DeleteInstance(ctx, "i-1234567890abcdef0")
		assert.NoError(t, err)
		assert.Equal(t, string(types.InstanceStateNameShuttingDown), state)
	})

	t.Run("error", func(t *testing.T) {
		mockClient := &mockEC2Client{}
		svc := NewService(mockClient)

		mockClient.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
			return nil, fmt.Errorf("API error")
		}

		ctx := context.Background()
		state, err := svc.DeleteInstance(ctx, "i-1234567890abcdef0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
		assert.Empty(t, state)
	})
}
