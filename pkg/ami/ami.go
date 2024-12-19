package ami

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2ClientAPI defines the interface for EC2 client operations
type EC2ClientAPI interface {
	DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	CreateSnapshot(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error)
	TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
	RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
	StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
	StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error)
	CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
}

// Service provides AMI management operations
type Service struct {
	client EC2ClientAPI
}

// NewService creates a new AMI service
func NewService(client EC2ClientAPI) *Service {
	return &Service{
		client: client,
	}
}

// GetAMIWithTag gets an AMI by its tag
func (s *Service) GetAMIWithTag(ctx context.Context, tagKey, tagValue string) (string, error) {
	input := &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:" + tagKey),
				Values: []string{tagValue},
			},
		},
	}

	result, err := s.client.DescribeImages(ctx, input)
	if err != nil {
		return "", fmt.Errorf("describe images: %w", err)
	}

	if len(result.Images) == 0 {
		return "", nil
	}

	return aws.ToString(result.Images[0].ImageId), nil
}

// TagAMI tags an AMI with the specified key and value
func (s *Service) TagAMI(ctx context.Context, amiID, tagKey, tagValue string) error {
	input := &ec2.CreateTagsInput{
		Resources: []string{amiID},
		Tags: []types.Tag{
			{
				Key:   aws.String(tagKey),
				Value: aws.String(tagValue),
			},
		},
	}

	_, err := s.client.CreateTags(ctx, input)
	return err
}

// MigrateInstances migrates instances to new AMI if they have the enabled tag
func (s *Service) MigrateInstances(ctx context.Context, oldAMI, newAMI, enabledValue string) error {
	instances, err := s.fetchEnabledInstances(ctx, enabledValue)
	if err != nil {
		return fmt.Errorf("fetch instances: %w", err)
	}

	if len(instances) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	for _, instance := range instances {
		shouldMigrate, needsStart := s.shouldMigrateInstance(instance)
		if !shouldMigrate {
			continue
		}

		wg.Add(1)
		go func(inst types.Instance, start bool) {
			defer wg.Done()

			// If instance needs to be started
			if start && inst.State.Name != types.InstanceStateNameRunning {
				if err := s.startInstance(ctx, inst); err != nil {
					s.tagInstanceAsFailed(ctx, inst)
					return
				}
			}

			// Perform migration
			if err := s.upgradeInstance(ctx, newAMI, inst); err != nil {
				s.tagInstanceAsFailed(ctx, inst)
				return
			}

			// If we started the instance, stop it again
			if start && inst.State.Name != types.InstanceStateNameRunning {
				if err := s.stopInstance(ctx, inst); err != nil {
					s.tagInstanceAsFailed(ctx, inst)
				}
			}
		}(instance, needsStart)
	}
	wg.Wait()

	return nil
}

func (s *Service) fetchEnabledInstances(ctx context.Context, enabledValue string) ([]types.Instance, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:ami-migrate"),
				Values: []string{enabledValue},
			},
		},
	}

	resp, err := s.client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, err
	}

	var instances []types.Instance
	for _, reservation := range resp.Reservations {
		instances = append(instances, reservation.Instances...)
	}
	return instances, nil
}

func (s *Service) shouldMigrateInstance(instance types.Instance) (bool, bool) {
	isRunning := instance.State.Name == types.InstanceStateNameRunning
	shouldStart := false

	// Check for if-running tag
	for _, tag := range instance.Tags {
		if aws.ToString(tag.Key) == "ami-migrate-if-running" &&
			aws.ToString(tag.Value) == "enabled" {
			shouldStart = true
			break
		}
	}

	// Migrate if running or has if-running tag
	return isRunning || shouldStart, shouldStart
}

func (s *Service) startInstance(ctx context.Context, instance types.Instance) error {
	input := &ec2.StartInstancesInput{
		InstanceIds: []string{aws.ToString(instance.InstanceId)},
	}
	_, err := s.client.StartInstances(ctx, input)
	if err != nil {
		return err
	}

	// Wait for instance to start
	waiter := ec2.NewInstanceRunningWaiter(s.client)
	return waiter.Wait(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{aws.ToString(instance.InstanceId)},
	}, 5*time.Minute)
}

func (s *Service) stopInstance(ctx context.Context, instance types.Instance) error {
	input := &ec2.StopInstancesInput{
		InstanceIds: []string{aws.ToString(instance.InstanceId)},
	}
	_, err := s.client.StopInstances(ctx, input)
	if err != nil {
		return err
	}

	// Wait for instance to stop
	waiter := ec2.NewInstanceStoppedWaiter(s.client)
	return waiter.Wait(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{aws.ToString(instance.InstanceId)},
	}, 5*time.Minute)
}

func (s *Service) upgradeInstance(ctx context.Context, newAMI string, instance types.Instance) error {
	// Create snapshot of the instance's volumes
	for _, mapping := range instance.BlockDeviceMappings {
		if mapping.Ebs != nil {
			_, err := s.client.CreateSnapshot(ctx, &ec2.CreateSnapshotInput{
				VolumeId: mapping.Ebs.VolumeId,
			})
			if err != nil {
				return fmt.Errorf("create snapshot: %w", err)
			}
		}
	}

	// Terminate the old instance
	_, err := s.client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{aws.ToString(instance.InstanceId)},
	})
	if err != nil {
		return fmt.Errorf("terminate instance: %w", err)
	}

	// Launch new instance with the new AMI
	_, err = s.client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:      aws.String(newAMI),
		InstanceType: instance.InstanceType,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	})
	if err != nil {
		return fmt.Errorf("run instance: %w", err)
	}

	return nil
}

func (s *Service) tagInstanceAsFailed(ctx context.Context, instance types.Instance) {
	input := &ec2.CreateTagsInput{
		Resources: []string{aws.ToString(instance.InstanceId)},
		Tags: []types.Tag{
			{
				Key:   aws.String("ami-migrate-status"),
				Value: aws.String("failed"),
			},
		},
	}
	_, _ = s.client.CreateTags(ctx, input)
}
