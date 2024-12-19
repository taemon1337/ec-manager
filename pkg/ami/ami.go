package ami

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2ClientAPI defines the interface for EC2 client operations
type EC2ClientAPI interface {
	DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
	CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	CreateSnapshot(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error)
	TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
	RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
}

// Service handles AMI-related operations
type Service struct {
	client EC2ClientAPI
}

// NewService creates a new AMI service
func NewService(client EC2ClientAPI) *Service {
	return &Service{client: client}
}

// GetAMIWithTag retrieves an AMI ID based on a tag
func (s *Service) GetAMIWithTag(ctx context.Context, tag string) (string, error) {
	input := &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:Status"),
				Values: []string{tag},
			},
		},
		Owners: []string{"self"},
	}

	result, err := s.client.DescribeImages(ctx, input)
	if err != nil {
		return "", fmt.Errorf("describe images: %w", err)
	}

	if len(result.Images) == 0 {
		return "", fmt.Errorf("no AMI found with tag %s", tag)
	}

	return aws.ToString(result.Images[0].ImageId), nil
}

// UpdateAMITags updates the tags for old and new AMIs
func (s *Service) UpdateAMITags(ctx context.Context, oldAMI, newAMI string) error {
	// Remove "latest" tag from old AMI
	if err := s.updateTags(ctx, oldAMI, "previous"); err != nil {
		return fmt.Errorf("update old AMI tags: %w", err)
	}

	// Add "latest" tag to new AMI
	if err := s.updateTags(ctx, newAMI, "latest"); err != nil {
		return fmt.Errorf("update new AMI tags: %w", err)
	}

	return nil
}

func (s *Service) updateTags(ctx context.Context, amiID string, status string) error {
	input := &ec2.CreateTagsInput{
		Resources: []string{amiID},
		Tags: []types.Tag{
			{
				Key:   aws.String("Status"),
				Value: aws.String(status),
			},
		},
	}

	_, err := s.client.CreateTags(ctx, input)
	return err
}

// MigrateInstances migrates instances from old AMI to new AMI
func (s *Service) MigrateInstances(ctx context.Context, oldAMI, newAMI string) error {
	instances, err := s.fetchInstancesWithAMI(ctx, oldAMI)
	if err != nil {
		return fmt.Errorf("fetch instances: %w", err)
	}

	if len(instances) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	for _, instance := range instances {
		wg.Add(1)
		go func(inst types.Instance) {
			defer wg.Done()
			if err := s.upgradeInstance(ctx, newAMI, inst); err != nil {
				s.tagInstanceAsFailed(ctx, inst)
			}
		}(instance)
	}
	wg.Wait()

	return nil
}

func (s *Service) fetchInstancesWithAMI(ctx context.Context, ami string) ([]types.Instance, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("image-id"),
				Values: []string{ami},
			},
		},
	}

	result, err := s.client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, err
	}

	var instances []types.Instance
	for _, reservation := range result.Reservations {
		instances = append(instances, reservation.Instances...)
	}
	return instances, nil
}

func (s *Service) upgradeInstance(ctx context.Context, newAMI string, instance types.Instance) error {
	// Create snapshot of the instance's volumes
	for _, mapping := range instance.BlockDeviceMappings {
		if mapping.Ebs != nil {
			if _, err := s.createSnapshot(ctx, *mapping.Ebs.VolumeId, *instance.InstanceId); err != nil {
				return fmt.Errorf("create snapshot: %w", err)
			}
		}
	}

	// Terminate the old instance
	if err := s.terminateInstance(ctx, instance); err != nil {
		return fmt.Errorf("terminate instance: %w", err)
	}

	// Launch new instance with new AMI
	if err := s.launchInstance(ctx, newAMI, instance); err != nil {
		return fmt.Errorf("launch instance: %w", err)
	}

	return nil
}

func (s *Service) createSnapshot(ctx context.Context, volumeID, instanceID string) (string, error) {
	input := &ec2.CreateSnapshotInput{
		VolumeId: aws.String(volumeID),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeSnapshot,
				Tags: []types.Tag{
					{
						Key:   aws.String("InstanceID"),
						Value: aws.String(instanceID),
					},
				},
			},
		},
	}

	result, err := s.client.CreateSnapshot(ctx, input)
	if err != nil {
		return "", err
	}

	return *result.SnapshotId, nil
}

func (s *Service) terminateInstance(ctx context.Context, instance types.Instance) error {
	input := &ec2.TerminateInstancesInput{
		InstanceIds: []string{*instance.InstanceId},
	}

	_, err := s.client.TerminateInstances(ctx, input)
	return err
}

func (s *Service) launchInstance(ctx context.Context, newAMI string, oldInstance types.Instance) error {
	input := &ec2.RunInstancesInput{
		ImageId:      aws.String(newAMI),
		InstanceType: oldInstance.InstanceType,
		MinCount:    aws.Int32(1),
		MaxCount:    aws.Int32(1),
		SubnetId:    oldInstance.SubnetId,
	}

	_, err := s.client.RunInstances(ctx, input)
	return err
}

func (s *Service) tagInstanceAsFailed(ctx context.Context, instance types.Instance) {
	input := &ec2.CreateTagsInput{
		Resources: []string{*instance.InstanceId},
		Tags: []types.Tag{
			{
				Key:   aws.String("MigrationStatus"),
				Value: aws.String("Failed"),
			},
		},
	}

	_, _ = s.client.CreateTags(ctx, input)
}
