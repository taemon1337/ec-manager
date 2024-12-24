package mock

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// DescribeInstances implements EC2Client
func (m *MockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if m.DescribeInstancesFunc != nil {
		return m.DescribeInstancesFunc(ctx, params, optFns...)
	}

	// Mock instance not found for specific ID
	if len(params.InstanceIds) > 0 && params.InstanceIds[0] != "i-123" {
		return nil, fmt.Errorf("instance not found: %s", params.InstanceIds[0])
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
