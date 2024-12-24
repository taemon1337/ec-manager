package mock

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// NewVolumeAvailableWaiter returns a mock volume available waiter
func (m *MockEC2Client) NewVolumeAvailableWaiter() *ec2.VolumeAvailableWaiter {
	return &ec2.VolumeAvailableWaiter{}
}

// NewInstanceStoppedWaiter returns a mock stopped waiter
func (m *MockEC2Client) NewInstanceStoppedWaiter() *ec2.InstanceStoppedWaiter {
	return &ec2.InstanceStoppedWaiter{}
}

// NewInstanceRunningWaiter returns a mock running waiter
func (m *MockEC2Client) NewInstanceRunningWaiter() *ec2.InstanceRunningWaiter {
	return &ec2.InstanceRunningWaiter{}
}

// NewInstanceTerminatedWaiter returns a mock terminated waiter
func (m *MockEC2Client) NewInstanceTerminatedWaiter() *ec2.InstanceTerminatedWaiter {
	return &ec2.InstanceTerminatedWaiter{}
}
