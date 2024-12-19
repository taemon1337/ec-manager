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
