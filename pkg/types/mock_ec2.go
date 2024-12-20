package types

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// MockEC2Client is a mock implementation of EC2ClientAPI
type MockEC2Client struct {
	DescribeInstancesOutput *ec2.DescribeInstancesOutput
	DescribeInstancesError  error
	DescribeImagesOutput   *ec2.DescribeImagesOutput
	DescribeImagesError    error
	RunInstancesOutput     *ec2.RunInstancesOutput
	RunInstancesError      error
	StopInstancesOutput    *ec2.StopInstancesOutput
	StopInstancesError     error
	StartInstancesOutput   *ec2.StartInstancesOutput
	StartInstancesError    error
	CreateTagsOutput       *ec2.CreateTagsOutput
	CreateTagsError        error
	TerminateInstancesOutput *ec2.TerminateInstancesOutput
	TerminateInstancesError  error
	CreateSnapshotOutput    *ec2.CreateSnapshotOutput
	CreateSnapshotError     error
	DescribeSnapshotsOutput *ec2.DescribeSnapshotsOutput
	DescribeSnapshotsError  error
	CreateVolumeOutput     *ec2.CreateVolumeOutput
	CreateVolumeError      error
	DescribeVolumesOutput  *ec2.DescribeVolumesOutput
	DescribeVolumesError   error
	AttachVolumeOutput     *ec2.AttachVolumeOutput
	AttachVolumeError      error
}

func (m *MockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if m.DescribeInstancesError != nil {
		return nil, m.DescribeInstancesError
	}
	return m.DescribeInstancesOutput, nil
}

func (m *MockEC2Client) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	if m.DescribeImagesError != nil {
		return nil, m.DescribeImagesError
	}
	return m.DescribeImagesOutput, nil
}

func (m *MockEC2Client) RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	if m.RunInstancesError != nil {
		return nil, m.RunInstancesError
	}
	return m.RunInstancesOutput, nil
}

func (m *MockEC2Client) StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
	if m.StopInstancesError != nil {
		return nil, m.StopInstancesError
	}
	return m.StopInstancesOutput, nil
}

func (m *MockEC2Client) StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error) {
	if m.StartInstancesError != nil {
		return nil, m.StartInstancesError
	}
	return m.StartInstancesOutput, nil
}

func (m *MockEC2Client) CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	if m.CreateTagsError != nil {
		return nil, m.CreateTagsError
	}
	return m.CreateTagsOutput, nil
}

func (m *MockEC2Client) TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	if m.TerminateInstancesError != nil {
		return nil, m.TerminateInstancesError
	}
	return m.TerminateInstancesOutput, nil
}

func (m *MockEC2Client) CreateSnapshot(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error) {
	if m.CreateSnapshotError != nil {
		return nil, m.CreateSnapshotError
	}
	return m.CreateSnapshotOutput, nil
}

func (m *MockEC2Client) DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error) {
	if m.DescribeSnapshotsError != nil {
		return nil, m.DescribeSnapshotsError
	}
	return m.DescribeSnapshotsOutput, nil
}

func (m *MockEC2Client) CreateVolume(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error) {
	if m.CreateVolumeError != nil {
		return nil, m.CreateVolumeError
	}
	return m.CreateVolumeOutput, nil
}

func (m *MockEC2Client) DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error) {
	if m.DescribeVolumesError != nil {
		return nil, m.DescribeVolumesError
	}
	return m.DescribeVolumesOutput, nil
}

func (m *MockEC2Client) AttachVolume(ctx context.Context, params *ec2.AttachVolumeInput, optFns ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error) {
	if m.AttachVolumeError != nil {
		return nil, m.AttachVolumeError
	}
	return m.AttachVolumeOutput, nil
}
