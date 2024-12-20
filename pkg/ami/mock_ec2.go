package ami

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// mockEC2Client implements a mock EC2 client for testing
type mockEC2Client struct {
	err       error
	images    []types.Image
	instances []types.Instance
	instance  *types.Instance
	snapshots []types.Snapshot
	volumes   []types.Volume
}

func (m *mockEC2Client) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ec2.DescribeImagesOutput{Images: m.images}, nil
}

func (m *mockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{Instances: m.instances},
		},
	}, nil
}

func (m *mockEC2Client) RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.instance == nil {
		return &ec2.RunInstancesOutput{}, nil
	}
	return &ec2.RunInstancesOutput{
		Instances: []types.Instance{*m.instance},
	}, nil
}

func (m *mockEC2Client) TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ec2.TerminateInstancesOutput{}, nil
}

func (m *mockEC2Client) CreateSnapshot(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ec2.CreateSnapshotOutput{}, nil
}

func (m *mockEC2Client) CreateSnapshots(ctx context.Context, params *ec2.CreateSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotsOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	snapshots := make([]types.SnapshotInfo, len(m.snapshots))
	for i, s := range m.snapshots {
		snapshots[i] = types.SnapshotInfo{
			SnapshotId: s.SnapshotId,
			VolumeId:   s.VolumeId,
			State:      s.State,
		}
	}
	return &ec2.CreateSnapshotsOutput{Snapshots: snapshots}, nil
}

func (m *mockEC2Client) CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ec2.CreateTagsOutput{}, nil
}

func (m *mockEC2Client) DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ec2.DescribeSnapshotsOutput{Snapshots: m.snapshots}, nil
}

func (m *mockEC2Client) DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ec2.DescribeVolumesOutput{Volumes: m.volumes}, nil
}

func (m *mockEC2Client) CreateVolume(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ec2.CreateVolumeOutput{}, nil
}

func (m *mockEC2Client) AttachVolume(ctx context.Context, params *ec2.AttachVolumeInput, optFns ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ec2.AttachVolumeOutput{}, nil
}

func (m *mockEC2Client) StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ec2.StartInstancesOutput{}, nil
}

func (m *mockEC2Client) StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ec2.StopInstancesOutput{}, nil
}
