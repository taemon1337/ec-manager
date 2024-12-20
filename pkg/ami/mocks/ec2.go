package mocks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// MockEC2Client is a mock implementation of the EC2 client interface
type MockEC2Client struct {
	DescribeImagesOutput     *ec2.DescribeImagesOutput
	DescribeImagesError      error
	CreateTagsOutput         *ec2.CreateTagsOutput
	CreateTagsError          error
	DescribeInstancesOutput  *ec2.DescribeInstancesOutput
	DescribeInstancesError   error
	CreateSnapshotOutput     *ec2.CreateSnapshotOutput
	CreateSnapshotError      error
	TerminateInstancesOutput *ec2.TerminateInstancesOutput
	TerminateInstancesError  error
	RunInstancesOutput       *ec2.RunInstancesOutput
	RunInstancesError        error
	StopInstancesOutput      *ec2.StopInstancesOutput
	StopInstancesError       error
	StartInstancesOutput     *ec2.StartInstancesOutput
	StartInstancesError      error
	DescribeSnapshotsOutput  *ec2.DescribeSnapshotsOutput
	DescribeSnapshotsError   error
	CreateVolumeOutput       *ec2.CreateVolumeOutput
	CreateVolumeError        error
	DescribeVolumesOutput    *ec2.DescribeVolumesOutput
	DescribeVolumesError     error
	AttachVolumeOutput       *ec2.AttachVolumeOutput
	AttachVolumeError        error
}

func (m *MockEC2Client) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	return m.DescribeImagesOutput, m.DescribeImagesError
}

func (m *MockEC2Client) CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	return m.CreateTagsOutput, m.CreateTagsError
}

func (m *MockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return m.DescribeInstancesOutput, m.DescribeInstancesError
}

func (m *MockEC2Client) CreateSnapshot(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error) {
	return m.CreateSnapshotOutput, m.CreateSnapshotError
}

func (m *MockEC2Client) TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	return m.TerminateInstancesOutput, m.TerminateInstancesError
}

func (m *MockEC2Client) RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	return m.RunInstancesOutput, m.RunInstancesError
}

func (m *MockEC2Client) StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
	return m.StopInstancesOutput, m.StopInstancesError
}

func (m *MockEC2Client) StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error) {
	return m.StartInstancesOutput, m.StartInstancesError
}

func (m *MockEC2Client) DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error) {
	return m.DescribeSnapshotsOutput, m.DescribeSnapshotsError
}

func (m *MockEC2Client) CreateVolume(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error) {
	return m.CreateVolumeOutput, m.CreateVolumeError
}

func (m *MockEC2Client) DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error) {
	return m.DescribeVolumesOutput, m.DescribeVolumesError
}

func (m *MockEC2Client) AttachVolume(ctx context.Context, params *ec2.AttachVolumeInput, optFns ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error) {
	return m.AttachVolumeOutput, m.AttachVolumeError
}
