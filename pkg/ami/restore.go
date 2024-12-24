package ami

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// RestoreInstance restores an instance from a snapshot
func (s *Service) RestoreInstance(ctx context.Context, instanceID, snapshotID string) error {
	// Get instance
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}
	result, err := s.client.DescribeInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}
	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return fmt.Errorf("instance not found: %s", instanceID)
	}
	instance := result.Reservations[0].Instances[0]

	// Get snapshot
	snapInput := &ec2.DescribeSnapshotsInput{
		SnapshotIds: []string{snapshotID},
	}
	snapResult, err := s.client.DescribeSnapshots(ctx, snapInput)
	if err != nil {
		return fmt.Errorf("failed to get snapshot: %w", err)
	}
	if len(snapResult.Snapshots) == 0 {
		return fmt.Errorf("snapshot not found: %s", snapshotID)
	}
	snapshot := snapResult.Snapshots[0]

	// Create volume from snapshot
	createVolumeInput := &ec2.CreateVolumeInput{
		AvailabilityZone: instance.Placement.AvailabilityZone,
		SnapshotId:       aws.String(snapshotID),
		VolumeType:       types.VolumeTypeGp2, // Use GP2 by default
	}
	volume, err := s.client.CreateVolume(ctx, createVolumeInput)
	if err != nil {
		return fmt.Errorf("failed to create volume: %w", err)
	}

	// Get device name from snapshot tags
	var deviceName string
	for _, tag := range snapshot.Tags {
		if aws.ToString(tag.Key) == "ami-migrate-device" {
			deviceName = aws.ToString(tag.Value)
			break
		}
	}
	if deviceName == "" {
		deviceName = "/dev/xvdf" // default device if not found
	}

	// Attach volume
	attachInput := &ec2.AttachVolumeInput{
		Device:     aws.String(deviceName),
		InstanceId: aws.String(instanceID),
		VolumeId:   volume.VolumeId,
	}
	if _, err := s.client.AttachVolume(ctx, attachInput); err != nil {
		return fmt.Errorf("failed to attach volume: %w", err)
	}

	fmt.Printf("Successfully restored volume from snapshot %s to instance %s at device %s\n", snapshotID, instanceID, deviceName)
	return nil
}
