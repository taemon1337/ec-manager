package waiters

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/stretchr/testify/mock"
)

// MockInstanceStoppedWaiter is a mock implementation of ec2.InstanceStoppedWaiter
type MockInstanceStoppedWaiter struct {
	mock.Mock
}

// Wait mocks the Wait method
func (m *MockInstanceStoppedWaiter) Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceStoppedWaiterOptions)) error {
	args := m.Called(ctx, params, maxWaitDur, optFns)
	return args.Error(0)
}

// MockInstanceRunningWaiter is a mock implementation of ec2.InstanceRunningWaiter
type MockInstanceRunningWaiter struct {
	mock.Mock
}

// Wait mocks the Wait method
func (m *MockInstanceRunningWaiter) Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceRunningWaiterOptions)) error {
	args := m.Called(ctx, params, maxWaitDur, optFns)
	return args.Error(0)
}

// MockInstanceTerminatedWaiter is a mock implementation of ec2.InstanceTerminatedWaiter
type MockInstanceTerminatedWaiter struct {
	mock.Mock
}

// Wait implements the waiter interface
func (m *MockInstanceTerminatedWaiter) Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceTerminatedWaiterOptions)) error {
	args := m.Called(ctx, params, maxWaitDur, optFns)
	return args.Error(0)
}

// MockVolumeAvailableWaiter is a mock implementation of ec2.VolumeAvailableWaiter
type MockVolumeAvailableWaiter struct {
	mock.Mock
}

// Wait implements the waiter interface
func (m *MockVolumeAvailableWaiter) Wait(ctx context.Context, params *ec2.DescribeVolumesInput, maxWaitDur time.Duration, optFns ...func(*ec2.VolumeAvailableWaiterOptions)) error {
	args := m.Called(ctx, params, maxWaitDur, optFns)
	return args.Error(0)
}
