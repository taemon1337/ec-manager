package types

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
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

	// Functions
	DescribeInstancesFunc  func(context.Context, *ec2.DescribeInstancesInput, ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	StopInstancesFunc      func(context.Context, *ec2.StopInstancesInput, ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
	StartInstancesFunc     func(context.Context, *ec2.StartInstancesInput, ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error)
	CreateImageFunc        func(context.Context, *ec2.CreateImageInput, ...func(*ec2.Options)) (*ec2.CreateImageOutput, error)
	DescribeImagesFunc     func(context.Context, *ec2.DescribeImagesInput, ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
	CreateTagsFunc         func(context.Context, *ec2.CreateTagsInput, ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	RunInstancesFunc       func(context.Context, *ec2.RunInstancesInput, ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
	TerminateInstancesFunc func(context.Context, *ec2.TerminateInstancesInput, ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
	AttachVolumeFunc       func(context.Context, *ec2.AttachVolumeInput, ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error)
	CreateSnapshotFunc     func(context.Context, *ec2.CreateSnapshotInput, ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error)
	CreateVolumeFunc       func(context.Context, *ec2.CreateVolumeInput, ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error)
	DescribeSnapshotsFunc  func(context.Context, *ec2.DescribeSnapshotsInput, ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error)
	DescribeVolumesFunc    func(context.Context, *ec2.DescribeVolumesInput, ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error)
}

// NewMockEC2Client creates a new mock EC2 client
func NewMockEC2Client() *MockEC2Client {
	return &MockEC2Client{}
}

// DescribeInstances implements EC2Client
func (m *MockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if m.DescribeInstancesFunc != nil {
		return m.DescribeInstancesFunc(ctx, params, optFns...)
	}
	return m.DescribeInstancesOutput, nil
}

// StopInstances implements EC2Client
func (m *MockEC2Client) StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
	if m.StopInstancesFunc != nil {
		return m.StopInstancesFunc(ctx, params, optFns...)
	}
	return m.StopInstancesOutput, nil
}

// StartInstances implements EC2Client
func (m *MockEC2Client) StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error) {
	if m.StartInstancesFunc != nil {
		return m.StartInstancesFunc(ctx, params, optFns...)
	}
	return m.StartInstancesOutput, nil
}

// CreateImage implements EC2Client
func (m *MockEC2Client) CreateImage(ctx context.Context, params *ec2.CreateImageInput, optFns ...func(*ec2.Options)) (*ec2.CreateImageOutput, error) {
	if m.CreateImageFunc != nil {
		return m.CreateImageFunc(ctx, params, optFns...)
	}
	return m.CreateImageOutput, nil
}

// DescribeImages implements EC2Client
func (m *MockEC2Client) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	if m.DescribeImagesFunc != nil {
		return m.DescribeImagesFunc(ctx, params, optFns...)
	}
	return m.DescribeImagesOutput, nil
}

// CreateTags implements EC2Client
func (m *MockEC2Client) CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	if m.CreateTagsFunc != nil {
		return m.CreateTagsFunc(ctx, params, optFns...)
	}
	return m.CreateTagsOutput, nil
}

// RunInstances implements EC2Client
func (m *MockEC2Client) RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	if m.RunInstancesFunc != nil {
		return m.RunInstancesFunc(ctx, params, optFns...)
	}
	return m.RunInstancesOutput, nil
}

// TerminateInstances implements EC2Client
func (m *MockEC2Client) TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	if m.TerminateInstancesFunc != nil {
		return m.TerminateInstancesFunc(ctx, params, optFns...)
	}
	return m.TerminateInstancesOutput, nil
}

// AttachVolume implements EC2Client
func (m *MockEC2Client) AttachVolume(ctx context.Context, params *ec2.AttachVolumeInput, optFns ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error) {
	if m.AttachVolumeFunc != nil {
		return m.AttachVolumeFunc(ctx, params, optFns...)
	}
	return m.AttachVolumeOutput, nil
}

// CreateSnapshot implements EC2Client
func (m *MockEC2Client) CreateSnapshot(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error) {
	if m.CreateSnapshotFunc != nil {
		return m.CreateSnapshotFunc(ctx, params, optFns...)
	}
	return m.CreateSnapshotOutput, nil
}

// CreateVolume implements EC2Client
func (m *MockEC2Client) CreateVolume(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error) {
	if m.CreateVolumeFunc != nil {
		return m.CreateVolumeFunc(ctx, params, optFns...)
	}
	return m.CreateVolumeOutput, nil
}

// DescribeSnapshots implements EC2Client
func (m *MockEC2Client) DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error) {
	if m.DescribeSnapshotsFunc != nil {
		return m.DescribeSnapshotsFunc(ctx, params, optFns...)
	}
	return m.DescribeSnapshotsOutput, nil
}

// DescribeVolumes implements EC2Client
func (m *MockEC2Client) DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error) {
	if m.DescribeVolumesFunc != nil {
		return m.DescribeVolumesFunc(ctx, params, optFns...)
	}
	return m.DescribeVolumesOutput, nil
}

// MockInstanceRunningWaiter is a mock implementation of instance running waiter
type MockInstanceRunningWaiter struct {
	*ec2.InstanceRunningWaiter
}

// Wait implements waiter interface
func (m *MockInstanceRunningWaiter) Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration) error {
	return nil
}

// MockInstanceStoppedWaiter is a mock implementation of instance stopped waiter
type MockInstanceStoppedWaiter struct {
	*ec2.InstanceStoppedWaiter
}

// Wait implements waiter interface
func (m *MockInstanceStoppedWaiter) Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration) error {
	return nil
}

// MockInstanceTerminatedWaiter is a mock implementation of instance terminated waiter
type MockInstanceTerminatedWaiter struct {
	*ec2.InstanceTerminatedWaiter
}

// Wait implements waiter interface
func (m *MockInstanceTerminatedWaiter) Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration) error {
	return nil
}

// NewInstanceRunningWaiter returns a mock running waiter
func (m *MockEC2Client) NewInstanceRunningWaiter() *MockInstanceRunningWaiter {
	return &MockInstanceRunningWaiter{}
}

// NewInstanceStoppedWaiter returns a mock stopped waiter
func (m *MockEC2Client) NewInstanceStoppedWaiter() *MockInstanceStoppedWaiter {
	return &MockInstanceStoppedWaiter{}
}

// NewInstanceTerminatedWaiter returns a mock terminated waiter
func (m *MockEC2Client) NewInstanceTerminatedWaiter() *MockInstanceTerminatedWaiter {
	return &MockInstanceTerminatedWaiter{}
}
