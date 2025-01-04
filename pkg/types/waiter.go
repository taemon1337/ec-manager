package types

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// WaiterAPI defines the common interface for all waiters
type WaiterAPI interface {
	Wait(ctx context.Context, params interface{}, maxWaitDur time.Duration, optFns ...interface{}) error
}

// InstanceRunningWaiterAPI defines the interface for instance running waiter
type InstanceRunningWaiterAPI interface {
	Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceRunningWaiterOptions)) error
}

// InstanceStoppedWaiterAPI defines the interface for instance stopped waiter
type InstanceStoppedWaiterAPI interface {
	Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceStoppedWaiterOptions)) error
}

// InstanceTerminatedWaiterAPI defines the interface for instance terminated waiter
type InstanceTerminatedWaiterAPI interface {
	Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceTerminatedWaiterOptions)) error
}

// VolumeAvailableWaiterAPI defines the interface for volume available waiter
type VolumeAvailableWaiterAPI interface {
	Wait(ctx context.Context, params *ec2.DescribeVolumesInput, maxWaitDur time.Duration, optFns ...func(*ec2.VolumeAvailableWaiterOptions)) error
}
