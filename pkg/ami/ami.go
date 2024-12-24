package ami

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2Client interface defines the methods we need from AWS SDK
type EC2Client interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
	CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
	StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
	StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error)
	AttachVolume(ctx context.Context, params *ec2.AttachVolumeInput, optFns ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error)
	CreateSnapshot(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error)
	TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
	CreateVolume(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error)
	CreateImage(ctx context.Context, params *ec2.CreateImageInput, optFns ...func(*ec2.Options)) (*ec2.CreateImageOutput, error)
	DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error)
	DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error)
	DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error)
	DescribeKeyPairs(ctx context.Context, params *ec2.DescribeKeyPairsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeKeyPairsOutput, error)
	NewInstanceRunningWaiter() *ec2.InstanceRunningWaiter
	NewInstanceStoppedWaiter() *ec2.InstanceStoppedWaiter
	NewInstanceTerminatedWaiter() *ec2.InstanceTerminatedWaiter
	NewVolumeAvailableWaiter() *ec2.VolumeAvailableWaiter
}

// InstanceConfig holds configuration for creating a new EC2 instance
type InstanceConfig struct {
	ImageID      string
	InstanceType string
	KeyName      string
	SubnetID     string
	UserData     string
}

// Service provides methods for managing EC2 instances
type Service struct {
	client EC2Client
}

// NewService creates a new Service instance
func NewService(client EC2Client) *Service {
	return &Service{
		client: client,
	}
}

// BackupInstance creates a backup AMI of the given instance
func (s *Service) BackupInstance(ctx context.Context, instanceID string) (string, error) {
	// First, describe the instance to make sure it exists
	describeOutput, err := s.client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe instance: %w", err)
	}

	if len(describeOutput.Reservations) == 0 || len(describeOutput.Reservations[0].Instances) == 0 {
		return "", fmt.Errorf("instance not found: %s", instanceID)
	}

	// Create an AMI from the instance
	createImageOutput, err := s.client.CreateImage(ctx, &ec2.CreateImageInput{
		InstanceId: aws.String(instanceID),
		Name:      aws.String(fmt.Sprintf("backup-%s-%s", instanceID, time.Now().Format("2006-01-02-15-04-05"))),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create AMI: %w", err)
	}

	// Tag the AMI
	_, err = s.client.CreateTags(ctx, &ec2.CreateTagsInput{
		Resources: []string{*createImageOutput.ImageId},
		Tags: []types.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(fmt.Sprintf("Backup of %s", instanceID)),
			},
			{
				Key:   aws.String("SourceInstanceId"),
				Value: aws.String(instanceID),
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to tag AMI: %w", err)
	}

	return *createImageOutput.ImageId, nil
}

// CreateInstance creates a new EC2 instance with the given configuration
func (s *Service) CreateInstance(ctx context.Context, cfg InstanceConfig) (string, error) {
	input := &ec2.RunInstancesInput{
		ImageId:      aws.String(cfg.ImageID),
		InstanceType: types.InstanceType(cfg.InstanceType),
		KeyName:      aws.String(cfg.KeyName),
		SubnetId:     aws.String(cfg.SubnetID),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	}

	if cfg.UserData != "" {
		input.UserData = aws.String(cfg.UserData)
	}

	output, err := s.client.RunInstances(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to create instance: %w", err)
	}

	if len(output.Instances) == 0 {
		return "", fmt.Errorf("no instance was created")
	}

	return *output.Instances[0].InstanceId, nil
}

// MigrateInstance migrates an EC2 instance to a new AMI
func (s *Service) MigrateInstance(ctx context.Context, instanceID string, newAMI string) (string, error) {
	// First, describe the instance to make sure it exists
	describeOutput, err := s.client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe instance: %w", err)
	}

	if len(describeOutput.Reservations) == 0 || len(describeOutput.Reservations[0].Instances) == 0 {
		return "", fmt.Errorf("instance not found: %s", instanceID)
	}

	instance := describeOutput.Reservations[0].Instances[0]

	// Create a new instance from the AMI with the same configuration
	cfg := InstanceConfig{
		ImageID:      newAMI,
		InstanceType: string(instance.InstanceType),
		KeyName:      *instance.KeyName,
		SubnetID:     *instance.SubnetId,
	}

	newInstanceID, err := s.CreateInstance(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create new instance: %w", err)
	}

	// Tag the new instance
	_, err = s.client.CreateTags(ctx, &ec2.CreateTagsInput{
		Resources: []string{newInstanceID},
		Tags: []types.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(fmt.Sprintf("Migrated from %s", instanceID)),
			},
			{
				Key:   aws.String("SourceInstanceId"),
				Value: aws.String(instanceID),
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to tag new instance: %w", err)
	}

	return newInstanceID, nil
}

// DeleteInstance terminates an EC2 instance
func (s *Service) DeleteInstance(ctx context.Context, instanceID string) (string, error) {
	// First, describe the instance to make sure it exists
	describeOutput, err := s.client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe instance: %w", err)
	}

	if len(describeOutput.Reservations) == 0 || len(describeOutput.Reservations[0].Instances) == 0 {
		return "", fmt.Errorf("instance not found: %s", instanceID)
	}

	// Terminate the instance
	terminateOutput, err := s.client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return "", fmt.Errorf("failed to terminate instance: %w", err)
	}

	if len(terminateOutput.TerminatingInstances) == 0 {
		return "", fmt.Errorf("instance was not terminated")
	}

	return string(terminateOutput.TerminatingInstances[0].CurrentState.Name), nil
}

// DescribeInstances returns a list of all EC2 instances
func (s *Service) DescribeInstances(ctx context.Context) (*ec2.DescribeInstancesOutput, error) {
	return s.client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
}

// ListSubnets returns a list of all VPC subnets
func (s *Service) ListSubnets(ctx context.Context) ([]types.Subnet, error) {
	output, err := s.client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe subnets: %w", err)
	}
	return output.Subnets, nil
}

// ListKeyPairs returns a list of all SSH key pairs
func (s *Service) ListKeyPairs(ctx context.Context) ([]types.KeyPairInfo, error) {
	output, err := s.client.DescribeKeyPairs(ctx, &ec2.DescribeKeyPairsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe key pairs: %w", err)
	}
	return output.KeyPairs, nil
}
