package ami

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Error constants
var (
	// Common errors
	ErrInstanceNotFound = errors.New("instance not found")
	ErrAMINotFound      = errors.New("AMI not found")
	ErrNoInstances      = errors.New("no instances launched")

	// Operation errors
	ErrRunInstances         = errors.New("failed to launch instance")
	ErrCreateTags           = errors.New("failed to create tags")
	ErrCreateImage          = errors.New("failed to create image")
	ErrCreateSnapshot       = errors.New("failed to create snapshot")
	ErrDescribeInstances    = errors.New("failed to describe instances")
	ErrDescribeImages       = errors.New("failed to describe images")
	ErrDescribeSnapshots    = errors.New("failed to describe snapshots")
	ErrStopInstance         = errors.New("failed to stop instance")
	ErrStartInstance        = errors.New("failed to start instance")
	ErrTerminateInstance    = errors.New("failed to terminate instance")
	ErrCreateImageFailed    = errors.New("failed to create image")
	ErrCreateImageNilOutput = errors.New("failed to create image: nil output")
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
	DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error)
	DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error)
	DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error)
	DescribeKeyPairs(ctx context.Context, params *ec2.DescribeKeyPairsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeKeyPairsOutput, error)
	CreateImage(ctx context.Context, params *ec2.CreateImageInput, optFns ...func(*ec2.Options)) (*ec2.CreateImageOutput, error)
	CreateVolume(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error)
	NewInstanceRunningWaiter() interface {
		Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceRunningWaiterOptions)) error
	}
	NewInstanceStoppedWaiter() interface {
		Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceStoppedWaiterOptions)) error
	}
	NewInstanceTerminatedWaiter() interface {
		Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration, optFns ...func(*ec2.InstanceTerminatedWaiterOptions)) error
	}
	NewVolumeAvailableWaiter() interface {
		Wait(ctx context.Context, params *ec2.DescribeVolumesInput, maxWaitDur time.Duration, optFns ...func(*ec2.VolumeAvailableWaiterOptions)) error
	}
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
		Name:       aws.String(fmt.Sprintf("backup-%s-%s", instanceID, time.Now().Format("2006-01-02-15-04-05"))),
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
	var tags []types.Tag
	tags = append(tags, types.Tag{
		Key:   aws.String("Name"),
		Value: aws.String(fmt.Sprintf("Migrated from %s", instanceID)),
	})
	tags = append(tags, types.Tag{
		Key:   aws.String("SourceInstanceId"),
		Value: aws.String(instanceID),
	})

	if len(tags) > 0 {
		tagInput := &ec2.CreateTagsInput{
			Resources: []string{newInstanceID},
			Tags:      tags,
		}
		_, err = s.client.CreateTags(ctx, tagInput)
		if err != nil {
			return "", fmt.Errorf("failed to create tags: %w", err)
		}
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

// GetInstance returns details of a specific EC2 instance
func (s *Service) GetInstance(ctx context.Context, instanceID string) (*types.Instance, error) {
	log.Printf("GetInstance: Looking up instance with ID: %s", instanceID)
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	output, err := s.client.DescribeInstances(ctx, input)
	if err != nil {
		log.Printf("GetInstance: Error from DescribeInstances: %v", err)
		return nil, fmt.Errorf("failed to describe instance: %w", err)
	}

	log.Printf("GetInstance: Got response with %d reservations", len(output.Reservations))
	if len(output.Reservations) == 0 {
		log.Printf("GetInstance: No reservations found for instance %s", instanceID)
		return nil, ErrInstanceNotFound
	}

	instances := output.Reservations[0].Instances
	log.Printf("GetInstance: First reservation has %d instances", len(instances))
	if len(instances) == 0 {
		log.Printf("GetInstance: No instances found in reservation for instance %s", instanceID)
		return nil, ErrInstanceNotFound
	}

	for i := range instances {
		if instances[i].InstanceId != nil && *instances[i].InstanceId == instanceID {
			instance := instances[i]
			return &instance, nil
		}
	}

	log.Printf("GetInstance: Instance %s not found in reservation", instanceID)
	return nil, ErrInstanceNotFound
}

// DescribeInstance returns details of a specific EC2 instance
func (s *Service) DescribeInstance(ctx context.Context, instanceID string) (*types.Instance, error) {
	output, err := s.client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instance: %w", err)
	}

	if len(output.Reservations) == 0 || len(output.Reservations[0].Instances) == 0 {
		return nil, nil
	}

	return &output.Reservations[0].Instances[0], nil
}

// DescribeImages returns details of specific AMIs
func (s *Service) DescribeImages(ctx context.Context, imageIDs []string) (*ec2.DescribeImagesOutput, error) {
	output, err := s.client.DescribeImages(ctx, &ec2.DescribeImagesInput{
		ImageIds: imageIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe images: %w", err)
	}

	return output, nil
}

// UpdateLatestAMITag removes the 'latest' tag from all AMIs with the given name prefix
// and adds it to the specified AMI
func (s *Service) UpdateLatestAMITag(ctx context.Context, namePrefix, newLatestAMI string) error {
	// Find all AMIs with the same prefix and latest tag
	input := &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("name"),
				Values: []string{namePrefix + "*"},
			},
			{
				Name:   aws.String("tag:ami-migrate"),
				Values: []string{"latest"},
			},
		},
	}

	output, err := s.client.DescribeImages(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to describe images: %w", err)
	}

	// Remove 'latest' tag from all existing AMIs
	for _, img := range output.Images {
		if *img.ImageId == newLatestAMI {
			continue // Skip the new AMI
		}
		_, err = s.client.CreateTags(ctx, &ec2.CreateTagsInput{
			Resources: []string{*img.ImageId},
			Tags: []types.Tag{
				{
					Key:   aws.String("ami-migrate"),
					Value: aws.String("outdated"),
				},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to update tags for AMI %s: %w", *img.ImageId, err)
		}
	}

	// Add 'latest' tag to the new AMI
	_, err = s.client.CreateTags(ctx, &ec2.CreateTagsInput{
		Resources: []string{newLatestAMI},
		Tags: []types.Tag{
			{
				Key:   aws.String("ami-migrate"),
				Value: aws.String("latest"),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to tag new AMI as latest: %w", err)
	}

	return nil
}

// GetLatestAMI returns the latest version of an AMI for the given OS
func (s *Service) GetLatestAMI(ctx context.Context, os string) (*types.Image, error) {
	input := &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:OS"),
				Values: []string{os},
			},
			{
				Name:   aws.String("tag:ami-migrate"),
				Values: []string{"latest"},
			},
		},
	}

	output, err := s.client.DescribeImages(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe images: %w", err)
	}

	if len(output.Images) == 0 {
		return nil, fmt.Errorf("no AMI found for OS %s", os)
	}

	return &output.Images[0], nil
}

// GetAMIByVersion returns a specific version of an AMI for the given OS
func (s *Service) GetAMIByVersion(ctx context.Context, os string, version string) (*types.Image, error) {
	input := &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:OS"),
				Values: []string{os},
			},
			{
				Name:   aws.String("tag:Version"),
				Values: []string{version},
			},
		},
	}

	output, err := s.client.DescribeImages(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe images: %w", err)
	}

	if len(output.Images) == 0 {
		return nil, fmt.Errorf("no AMI found for OS %s version %s", os, version)
	}

	return &output.Images[0], nil
}

// GetInstanceOS returns the OS of an instance based on its AMI
func (s *Service) GetInstanceOS(ctx context.Context, instanceID string) (string, error) {
	instance, err := s.GetInstance(ctx, instanceID)
	if err != nil {
		return "", err
	}

	if instance.ImageId == nil {
		return "", fmt.Errorf("instance %s has no AMI ID", instanceID)
	}

	input := &ec2.DescribeImagesInput{
		ImageIds: []string{*instance.ImageId},
	}

	output, err := s.client.DescribeImages(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to describe AMI: %w", err)
	}

	if len(output.Images) == 0 {
		return "", fmt.Errorf("AMI %s not found", *instance.ImageId)
	}

	for _, tag := range output.Images[0].Tags {
		if *tag.Key == "OS" {
			return *tag.Value, nil
		}
	}

	return "", fmt.Errorf("AMI %s has no OS tag", *instance.ImageId)
}

// UpdateAMITags updates the tags of an AMI
func (s *Service) UpdateAMITags(ctx context.Context, amiID string, tags map[string]string) error {
	var ec2Tags []types.Tag
	for key, value := range tags {
		ec2Tags = append(ec2Tags, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	_, err := s.client.CreateTags(ctx, &ec2.CreateTagsInput{
		Resources: []string{amiID},
		Tags:      ec2Tags,
	})
	return err
}

// LaunchInstance launches a new EC2 instance from the given AMI ID
func (s *Service) LaunchInstance(ctx context.Context, amiID string, name string) (*types.Instance, error) {
	// Launch new instance
	runOutput, err := s.client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:      aws.String(amiID),
		InstanceType: types.InstanceTypeT2Micro,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRunInstances, err)
	}

	if len(runOutput.Instances) == 0 {
		return nil, ErrNoInstances
	}

	instance := runOutput.Instances[0]

	// Tag the instance
	var tags []types.Tag
	if name != "" {
		tags = append(tags, types.Tag{
			Key:   aws.String("Name"),
			Value: aws.String(name),
		})
	}

	if len(tags) > 0 {
		tagOutput, tagErr := s.client.CreateTags(ctx, &ec2.CreateTagsInput{
			Resources: []string{*instance.InstanceId},
			Tags:      tags,
		})
		if tagErr != nil {
			return nil, fmt.Errorf("%w: %v", ErrCreateTags, tagErr)
		}
		if tagOutput == nil {
			return nil, fmt.Errorf("%w: nil output", ErrCreateTags)
		}
	}

	return &instance, nil
}

// AMI represents an Amazon Machine Image
type AMI struct {
	ImageId      string
	Name         string
	Desc         string
	SourceAMI    string
	InstanceType string
	SubnetID     string
	service      *Service
}

// NewAMI creates a new AMI instance
func NewAMI(service *Service) *AMI {
	return &AMI{
		service: service,
	}
}

// Launch launches a new EC2 instance from this AMI
func (a *AMI) Launch(ctx context.Context) (*types.Instance, error) {
	return a.service.LaunchInstance(ctx, a.ImageId, a.Name)
}

// FindAMI finds an AMI by its ID
func (s *Service) FindAMI(ctx context.Context, amiID string) (*types.Image, error) {
	input := &ec2.DescribeImagesInput{
		ImageIds: []string{amiID},
	}

	output, err := s.client.DescribeImages(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe AMI: %w", err)
	}

	if output == nil || len(output.Images) == 0 {
		return nil, fmt.Errorf("AMI not found")
	}

	return &output.Images[0], nil
}

// GetImage gets an AMI by ID
func (s *Service) GetImage(ctx context.Context, imageId string) (*types.Image, error) {
	input := &ec2.DescribeImagesInput{
		ImageIds: []string{imageId},
	}

	output, err := s.client.DescribeImages(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe images: %w", err)
	}

	if output == nil || len(output.Images) == 0 {
		return nil, fmt.Errorf("AMI not found")
	}

	return &output.Images[0], nil
}

// CreateAMI creates a new AMI from an instance
func (s *Service) CreateAMI(ctx context.Context, instanceID, name, description string) (*types.Image, error) {
	log.Printf("CreateAMI: Starting AMI creation for instance %s with name %s", instanceID, name)

	// First, verify the instance exists
	instance, err := s.GetInstance(ctx, instanceID)
	if err != nil {
		log.Printf("CreateAMI: Failed to get instance: %v", err)
		return nil, err
	}
	log.Printf("CreateAMI: Successfully found instance %s", instanceID)

	// Create the AMI
	createOutput, err := s.client.CreateImage(ctx, &ec2.CreateImageInput{
		InstanceId:  instance.InstanceId,
		Name:        aws.String(name),
		Description: aws.String(description),
	})
	if err != nil {
		log.Printf("CreateAMI: Failed to create AMI: %v", err)
		return nil, fmt.Errorf("failed to create AMI: %w", err)
	}
	if createOutput == nil {
		log.Printf("CreateAMI: Failed to create AMI: nil output")
		return nil, fmt.Errorf("failed to create AMI: nil output")
	}
	log.Printf("CreateAMI: Successfully created AMI with ID: %s", *createOutput.ImageId)

	// Add tags to the AMI
	output, err := s.client.CreateTags(ctx, &ec2.CreateTagsInput{
		Resources: []string{*createOutput.ImageId},
		Tags: []types.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(name),
			},
		},
	})
	if err != nil {
		log.Printf("CreateAMI: Failed to add tags to AMI %s: %v", *createOutput.ImageId, err)
		return nil, fmt.Errorf("failed to add tags to AMI: %w", err)
	}
	if output == nil {
		log.Printf("CreateAMI: Failed to add tags to AMI %s: nil output", *createOutput.ImageId)
		return nil, fmt.Errorf("failed to add tags to AMI: nil output")
	}
	log.Printf("CreateAMI: Successfully added tags to AMI %s", *createOutput.ImageId)

	return &types.Image{
		ImageId: createOutput.ImageId,
		Name:    aws.String(name),
	}, nil
}

// StopInstance stops an EC2 instance
func (s *Service) StopInstance(ctx context.Context, instanceID string) error {
	input := &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	}

	_, err := s.client.StopInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	// Poll for instance state
	for {
		describeInput := &ec2.DescribeInstancesInput{
			InstanceIds: []string{instanceID},
		}

		output, err := s.client.DescribeInstances(ctx, describeInput)
		if err != nil {
			return fmt.Errorf("failed to describe instance: %w", err)
		}

		if len(output.Reservations) == 0 || len(output.Reservations[0].Instances) == 0 {
			return fmt.Errorf("instance not found")
		}

		state := output.Reservations[0].Instances[0].State.Name
		if state == types.InstanceStateNameStopped {
			break
		}

		time.Sleep(5 * time.Second)
	}

	return nil
}

// StartInstance starts an EC2 instance and waits for it to be running
func (s *Service) StartInstance(ctx context.Context, instanceID string) error {
	// Start the instance
	_, err := s.client.StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}

	// Wait for the instance to be running
	waiter := s.client.NewInstanceRunningWaiter()
	err = waiter.Wait(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("error waiting for instance to start: %w", err)
	}

	// Verify the instance state
	_, err = s.client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return fmt.Errorf("error verifying instance state: %w", err)
	}

	return nil
}

// RestartInstance restarts an EC2 instance by stopping and starting it
func (s *Service) RestartInstance(ctx context.Context, instanceID string) error {
	// First stop the instance
	err := s.StopInstance(ctx, instanceID)
	if err != nil {
		return err
	}

	// Then start the instance
	err = s.StartInstance(ctx, instanceID)
	if err != nil {
		return err
	}

	return nil
}

// InstanceConfig holds configuration for creating a new EC2 instance
type InstanceConfig struct {
	ImageID      string
	InstanceType string
	KeyName      string
	SubnetID     string
	UserData     string
}
