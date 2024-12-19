package cmd

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// mockEC2Client implements ami.EC2ClientAPI for testing
type mockEC2Client struct {
	images    []types.Image
	instances []types.Instance
}

func (m *mockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	var filteredInstances []types.Instance
	if params.InstanceIds != nil && len(params.InstanceIds) > 0 {
		// Return the instance for specific instance ID
		for _, instance := range m.instances {
			if *instance.InstanceId == params.InstanceIds[0] {
				// Add volume information for backup tests
				instance.BlockDeviceMappings = []types.InstanceBlockDeviceMapping{
					{
						DeviceName: aws.String("/dev/sda1"),
						Ebs: &types.EbsInstanceBlockDevice{
							VolumeId: aws.String("vol-123"),
						},
					},
				}
				filteredInstances = append(filteredInstances, instance)
				break
			}
		}
	} else {
		// Filter instances by tags
		for _, instance := range m.instances {
			matches := true
			for _, filter := range params.Filters {
				if *filter.Name == "tag:ami-migrate" {
					matches = false
					for _, tag := range instance.Tags {
						if *tag.Key == "ami-migrate" && *tag.Value == filter.Values[0] {
							matches = true
							break
						}
					}
				}
			}
			if matches {
				// Add volume information for backup tests
				instance.BlockDeviceMappings = []types.InstanceBlockDeviceMapping{
					{
						DeviceName: aws.String("/dev/sda1"),
						Ebs: &types.EbsInstanceBlockDevice{
							VolumeId: aws.String("vol-123"),
						},
					},
				}
				filteredInstances = append(filteredInstances, instance)
			}
		}
	}
	return &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: filteredInstances,
			},
		},
	}, nil
}

func (m *mockEC2Client) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	var filteredImages []types.Image
	for _, image := range m.images {
		matches := true
		for _, filter := range params.Filters {
			if *filter.Name == "tag:Status" {
				matches = false
				for _, tag := range image.Tags {
					if *tag.Key == "Status" && *tag.Value == filter.Values[0] {
						matches = true
						break
					}
				}
			}
		}
		if matches {
			filteredImages = append(filteredImages, image)
		}
	}
	return &ec2.DescribeImagesOutput{
		Images: filteredImages,
	}, nil
}

func (m *mockEC2Client) CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	return &ec2.CreateTagsOutput{}, nil
}

func (m *mockEC2Client) CreateSnapshot(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error) {
	return &ec2.CreateSnapshotOutput{
		SnapshotId: aws.String("snap-123"),
	}, nil
}

func (m *mockEC2Client) StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
	return &ec2.StopInstancesOutput{}, nil
}

func (m *mockEC2Client) StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error) {
	return &ec2.StartInstancesOutput{}, nil
}

func (m *mockEC2Client) RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	instance := types.Instance{
		InstanceId:   aws.String("i-new"),
		ImageId:      params.ImageId,
		InstanceType: params.InstanceType,
		State: &types.InstanceState{
			Name: types.InstanceStateNameStopped,
		},
	}
	return &ec2.RunInstancesOutput{
		Instances: []types.Instance{instance},
	}, nil
}

func (m *mockEC2Client) TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	return &ec2.TerminateInstancesOutput{}, nil
}

func (m *mockEC2Client) DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error) {
	return &ec2.DescribeSnapshotsOutput{}, nil
}

func (m *mockEC2Client) CreateVolume(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error) {
	return &ec2.CreateVolumeOutput{}, nil
}

func (m *mockEC2Client) DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error) {
	return &ec2.DescribeVolumesOutput{}, nil
}

func (m *mockEC2Client) AttachVolume(ctx context.Context, params *ec2.AttachVolumeInput, optFns ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error) {
	return &ec2.AttachVolumeOutput{}, nil
}
