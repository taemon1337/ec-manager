package mock

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// AttachVolume implements EC2Client
func (m *MockEC2Client) AttachVolume(ctx context.Context, params *ec2.AttachVolumeInput, optFns ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error) {
	if m.AttachVolumeFunc != nil {
		return m.AttachVolumeFunc(ctx, params, optFns...)
	}
	fmt.Printf("Mock: Attaching volume %s to instance %s at device %s\n", *params.VolumeId, *params.InstanceId, *params.Device)
	return m.AttachVolumeOutput, nil
}

// CreateVolume implements EC2Client
func (m *MockEC2Client) CreateVolume(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error) {
	if m.CreateVolumeFunc != nil {
		return m.CreateVolumeFunc(ctx, params, optFns...)
	}
	fmt.Printf("Mock: Creating volume from snapshot %s in AZ %s\n", *params.SnapshotId, *params.AvailabilityZone)
	return m.CreateVolumeOutput, nil
}

// DescribeVolumes implements EC2Client
func (m *MockEC2Client) DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error) {
	if m.DescribeVolumesFunc != nil {
		return m.DescribeVolumesFunc(ctx, params, optFns...)
	}
	return m.DescribeVolumesOutput, nil
}
