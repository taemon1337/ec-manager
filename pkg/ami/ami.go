package ami

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/taemon1337/ec-manager/pkg/config"
	"github.com/taemon1337/ec-manager/pkg/logger"
	ecTypes "github.com/taemon1337/ec-manager/pkg/types"
)

var instanceStateWaiter waiterInterface

// Service provides AMI management operations
type Service struct {
	client ecTypes.EC2ClientAPI
}

// NewService creates a new AMI service
func NewService(client ecTypes.EC2ClientAPI) *Service {
	return &Service{
		client: client,
	}
}

// GetAMIWithTag gets an AMI by its tag
func (s *Service) GetAMIWithTag(ctx context.Context, tagKey, tagValue string) (*types.Image, error) {
	logger.Debug("Looking for AMI", "tagKey", tagKey, "tagValue", tagValue)

	input := &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", tagKey)),
				Values: []string{tagValue},
			},
		},
	}

	result, err := s.client.DescribeImages(ctx, input)
	if err != nil {
		logger.Error("Failed to describe images", "error", err)
		return nil, fmt.Errorf("describe images: %w", err)
	}

	if len(result.Images) == 0 {
		logger.Warn("No AMI found with tag", "tagKey", tagKey, "tagValue", tagValue)
		return nil, fmt.Errorf("no AMI found with tag %s=%s", tagKey, tagValue)
	}

	logger.Info("Found AMI", "amiID", *result.Images[0].ImageId)
	return &result.Images[0], nil
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
func (s *Service) MigrateInstances(ctx context.Context, enabledValue string) error {
	logger.Info("Starting migration of enabled instances", "enabledValue", enabledValue)

	// Get enabled instances
	instances, err := s.fetchEnabledInstances(ctx, enabledValue)
	if err != nil {
		logger.Error("Failed to fetch enabled instances", "error", err)
		return fmt.Errorf("fetch enabled instances: %w", err)
	}

	if len(instances) == 0 {
		logger.Info("No instances found with enabled tag")
		return nil
	}

	// Process instances concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, len(instances))

	for _, instance := range instances {
		wg.Add(1)
		go func(inst types.Instance) {
			defer wg.Done()

			// Get the OS type
			osType, err := s.GetInstanceOSType(ctx, aws.ToString(inst.InstanceId))
			if err != nil {
				errChan <- fmt.Errorf("get instance OS type %s: %w", aws.ToString(inst.InstanceId), err)
				return
			}

			// Get the latest AMI
			latestAMI, err := s.GetLatestAMI(ctx, osType)
			if err != nil {
				errChan <- fmt.Errorf("get latest AMI for instance %s: %w", aws.ToString(inst.InstanceId), err)
				return
			}

			if err := s.MigrateInstance(ctx, aws.ToString(inst.InstanceId), latestAMI); err != nil {
				errChan <- fmt.Errorf("migrate instance %s: %w", aws.ToString(inst.InstanceId), err)
			}
		}(instance)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)

	// Check for any errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to migrate some instances: %v", errs)
	}

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
	isRunning := string(instance.State.Name) == string(types.InstanceStateNameRunning)
	hasIfRunningTag := false

	// Check for if-running tag
	for _, tag := range instance.Tags {
		if aws.ToString(tag.Key) == "ami-migrate-if-running" &&
			aws.ToString(tag.Value) == "enabled" {
			hasIfRunningTag = true
			break
		}
	}

	// If instance is running, we need both tags
	if isRunning {
		return hasIfRunningTag, false
	}

	// If instance is stopped, we only need ami-migrate tag (which is already checked in fetchEnabledInstances)
	return true, false
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
	return waitForInstanceState(ctx, s.client, aws.ToString(instance.InstanceId), types.InstanceStateNameRunning)
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
	return waitForInstanceState(ctx, s.client, aws.ToString(instance.InstanceId), types.InstanceStateNameStopped)
}

func (s *Service) upgradeInstance(ctx context.Context, instance types.Instance, newAMI string) error {
	// Create snapshot of the instance's volumes
	for _, mapping := range instance.BlockDeviceMappings {
		if mapping.Ebs != nil {
			_, err := s.client.CreateSnapshot(ctx, &ec2.CreateSnapshotInput{
				VolumeId: mapping.Ebs.VolumeId,
				Description: aws.String(fmt.Sprintf("Backup before AMI migration for instance %s",
					aws.ToString(instance.InstanceId))),
			})
			if err != nil {
				return fmt.Errorf("create snapshot: %w", err)
			}
		}
	}

	// Stop the instance
	if string(instance.State.Name) == string(types.InstanceStateNameRunning) {
		if err := s.stopInstance(ctx, instance); err != nil {
			return fmt.Errorf("stop instance: %w", err)
		}
	}

	// Create new instance with new AMI
	runInput := &ec2.RunInstancesInput{
		ImageId:      aws.String(newAMI),
		InstanceType: instance.InstanceType,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	}

	runResult, err := s.client.RunInstances(ctx, runInput)
	if err != nil {
		return fmt.Errorf("run instances: %w", err)
	}

	// Terminate old instance
	_, err = s.client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{aws.ToString(instance.InstanceId)},
	})
	if err != nil {
		return fmt.Errorf("terminate instance: %w", err)
	}

	// Copy tags to new instance
	if err := s.copyTags(ctx, instance, runResult.Instances[0]); err != nil {
		return fmt.Errorf("copy tags: %w", err)
	}

	return nil
}

func (s *Service) copyTags(ctx context.Context, oldInstance, newInstance types.Instance) error {
	var tags []types.Tag
	for _, tag := range oldInstance.Tags {
		// Skip the migration status tag
		if aws.ToString(tag.Key) == "ami-migrate-status" {
			continue
		}
		tags = append(tags, tag)
	}

	input := &ec2.CreateTagsInput{
		Resources: []string{aws.ToString(newInstance.InstanceId)},
		Tags:      tags,
	}

	_, err := s.client.CreateTags(ctx, input)
	return err
}

func (s *Service) tagInstanceStatus(ctx context.Context, instance types.Instance, status, message string) error {
	input := &ec2.CreateTagsInput{
		Resources: []string{aws.ToString(instance.InstanceId)},
		Tags: []types.Tag{
			{
				Key:   aws.String("ami-migrate-status"),
				Value: aws.String(status),
			},
			{
				Key:   aws.String("ami-migrate-message"),
				Value: aws.String(message),
			},
			{
				Key:   aws.String("ami-migrate-timestamp"),
				Value: aws.String(time.Now().UTC().Format(time.RFC3339)),
			},
		},
	}

	_, err := s.client.CreateTags(ctx, input)
	return err
}

func (s *Service) BackupInstances(ctx context.Context, enabledValue string) error {
	// Get instances with ami-migrate tag
	instances, err := s.getInstances(ctx, enabledValue)
	if err != nil {
		return fmt.Errorf("failed to get instances: %w", err)
	}

	for _, instance := range instances {
		// Check if instance should be backed up based on state
		if string(instance.State.Name) == string(types.InstanceStateNameRunning) {
			// Check if running instance has the required tag
			if !hasTag(instance.Tags, "ami-migrate-if-running", enabledValue) {
				s.tagInstanceStatus(ctx, instance, "skipped", "Running instance without ami-migrate-if-running tag")
				continue
			}
		}

		s.tagInstanceStatus(ctx, instance, "in-progress", "Creating volume snapshots")

		// Create snapshots for each volume
		for _, device := range instance.BlockDeviceMappings {
			if device.Ebs == nil {
				continue
			}

			description := fmt.Sprintf("Backup of volume %s from instance %s",
				aws.ToString(device.Ebs.VolumeId),
				aws.ToString(instance.InstanceId))

			input := &ec2.CreateSnapshotInput{
				VolumeId:    device.Ebs.VolumeId,
				Description: aws.String(description),
				TagSpecifications: []types.TagSpecification{
					{
						ResourceType: types.ResourceTypeSnapshot,
						Tags: []types.Tag{
							{
								Key:   aws.String("ami-migrate-instance"),
								Value: instance.InstanceId,
							},
							{
								Key:   aws.String("ami-migrate-device"),
								Value: device.DeviceName,
							},
						},
					},
				},
			}

			_, err := s.client.CreateSnapshot(ctx, input)
			if err != nil {
				s.tagInstanceStatus(ctx, instance, "failed", fmt.Sprintf("Failed to create snapshot: %v", err))
				return fmt.Errorf("failed to create snapshot: %w", err)
			}
		}

		s.tagInstanceStatus(ctx, instance, "completed", "Volume snapshots created successfully")
	}

	return nil
}

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

	// Wait for volume to be available
	waiter := ec2.NewVolumeAvailableWaiter(s.client)
	if err := waiter.Wait(ctx, &ec2.DescribeVolumesInput{
		VolumeIds: []string{aws.ToString(volume.VolumeId)},
	}, config.GetTimeout()); err != nil {
		return fmt.Errorf("volume did not become available: %w", err)
	}

	// Stop instance if running
	if string(instance.State.Name) == string(types.InstanceStateNameRunning) {
		stopInput := &ec2.StopInstancesInput{
			InstanceIds: []string{instanceID},
		}
		if _, err := s.client.StopInstances(ctx, stopInput); err != nil {
			return fmt.Errorf("failed to stop instance: %w", err)
		}

		// Wait for instance to stop
		stopWaiter := ec2.NewInstanceStoppedWaiter(s.client)
		if err := stopWaiter.Wait(ctx, &ec2.DescribeInstancesInput{
			InstanceIds: []string{instanceID},
		}, config.GetTimeout()); err != nil {
			return fmt.Errorf("instance did not stop: %w", err)
		}
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

	return nil
}

func (s *Service) getInstances(ctx context.Context, enabledValue string) ([]types.Instance, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:ami-migrate"),
				Values: []string{enabledValue},
			},
		},
	}

	result, err := s.client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("describe instances: %w", err)
	}

	var instances []types.Instance
	for _, reservation := range result.Reservations {
		instances = append(instances, reservation.Instances...)
	}

	return instances, nil
}

func (s *Service) MigrateInstance(ctx context.Context, instanceID string, newAMI string) error {
	logger.Info("Starting instance migration", "instanceID", instanceID, "newAMI", newAMI)

	// Get the instance
	instance, err := s.getInstance(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("get instance: %w", err)
	}

	// Get the current AMI ID
	currentAMI := aws.ToString(instance.ImageId)
	if currentAMI == newAMI {
		return nil // Already on target AMI
	}

	// Perform the migration
	return s.migrateInstanceToAMI(ctx, instance, newAMI)
}

func (s *Service) GetLatestAMI(ctx context.Context, osType string) (string, error) {
	input := &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:ami-migrate"),
				Values: []string{"latest"},
			},
			{
				Name:   aws.String("tag:OS"),
				Values: []string{osType},
			},
		},
	}

	result, err := s.client.DescribeImages(ctx, input)
	if err != nil {
		return "", fmt.Errorf("describe images: %w", err)
	}

	if len(result.Images) == 0 {
		return "", fmt.Errorf("no AMI found for OS type: %s", osType)
	}

	// Sort images by creation date to get the most recent one
	latestImage := result.Images[0]
	for _, image := range result.Images[1:] {
		if aws.ToString(image.CreationDate) > aws.ToString(latestImage.CreationDate) {
			latestImage = image
		}
	}

	return aws.ToString(latestImage.ImageId), nil
}

func (s *Service) GetInstanceOSType(ctx context.Context, instanceID string) (string, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	result, err := s.client.DescribeInstances(ctx, input)
	if err != nil {
		return "", fmt.Errorf("describe instance: %w", err)
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return "", fmt.Errorf("instance not found: %s", instanceID)
	}

	instance := result.Reservations[0].Instances[0]

	// First check platform details
	if instance.PlatformDetails != nil {
		details := aws.ToString(instance.PlatformDetails)
		switch {
		case strings.Contains(details, "Red Hat"):
			return "RHEL9", nil
		case strings.Contains(details, "Ubuntu"):
			return "Ubuntu", nil
		}
	}

	// If platform details don't help, check the AMI details
	if instance.ImageId != nil {
		imageInput := &ec2.DescribeImagesInput{
			ImageIds: []string{aws.ToString(instance.ImageId)},
		}
		imageResult, err := s.client.DescribeImages(ctx, imageInput)
		if err == nil && len(imageResult.Images) > 0 {
			image := imageResult.Images[0]
			name := aws.ToString(image.Name)
			description := aws.ToString(image.Description)

			switch {
			case strings.Contains(strings.ToLower(name), "rhel") ||
				strings.Contains(strings.ToLower(description), "red hat"):
				return "RHEL9", nil
			case strings.Contains(strings.ToLower(name), "ubuntu") ||
				strings.Contains(strings.ToLower(description), "ubuntu"):
				return "Ubuntu", nil
			}
		}
	}

	// Finally, check instance tags as a fallback
	for _, tag := range instance.Tags {
		if aws.ToString(tag.Key) == "OS" {
			return aws.ToString(tag.Value), nil
		}
	}

	return "", fmt.Errorf("unable to determine OS type for instance: %s", instanceID)
}

func (s *Service) getInstance(ctx context.Context, instanceID string) (types.Instance, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	result, err := s.client.DescribeInstances(ctx, input)
	if err != nil {
		return types.Instance{}, fmt.Errorf("describe instance: %w", err)
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return types.Instance{}, fmt.Errorf("instance not found: %s", instanceID)
	}

	return result.Reservations[0].Instances[0], nil
}

func (s *Service) migrateInstanceToAMI(ctx context.Context, instance types.Instance, newAMI string) error {
	// Tag the instance to indicate migration is in progress
	err := s.tagInstanceStatus(ctx, instance, "migrating", fmt.Sprintf("Migrating to AMI: %s", newAMI))
	if err != nil {
		return fmt.Errorf("tag instance status: %w", err)
	}

	// Stop the instance if it's running
	if instance.State != nil && instance.State.Name == types.InstanceStateNameRunning {
		if err := s.stopInstance(ctx, instance); err != nil {
			return fmt.Errorf("stop instance: %w", err)
		}
	}

	// Perform the upgrade
	if err := s.upgradeInstance(ctx, instance, newAMI); err != nil {
		s.tagInstanceStatus(ctx, instance, "failed", fmt.Sprintf("Migration failed: %v", err))
		return fmt.Errorf("upgrade instance: %w", err)
	}

	// Tag the instance as successfully migrated
	return s.tagInstanceStatus(ctx, instance, "completed", fmt.Sprintf("Migrated to AMI: %s", newAMI))
}

func (s *Service) BackupInstance(ctx context.Context, instanceID string) error {
	logger.Info("Starting instance backup", "instanceID", instanceID)

	// Get instance details
	instance, err := s.getInstance(ctx, instanceID)
	if err != nil {
		logger.Error("Failed to get instance", "instanceID", instanceID, "error", err)
		return fmt.Errorf("failed to get instance: %v", err)
	}

	// Create snapshots for each volume
	for _, blockDevice := range instance.BlockDeviceMappings {
		if blockDevice.Ebs != nil {
			volumeID := aws.ToString(blockDevice.Ebs.VolumeId)
			deviceName := aws.ToString(blockDevice.DeviceName)
			logger.Debug("Creating snapshot for volume", "instanceID", instanceID, "volumeID", volumeID, "deviceName", deviceName)

			input := &ec2.CreateSnapshotInput{
				VolumeId:    aws.String(volumeID),
				Description: aws.String(fmt.Sprintf("Backup of volume %s from instance %s", volumeID, instanceID)),
				TagSpecifications: []types.TagSpecification{
					{
						ResourceType: types.ResourceTypeSnapshot,
						Tags: []types.Tag{
							{
								Key:   aws.String("Name"),
								Value: aws.String(fmt.Sprintf("Backup-%s-%s", instanceID, time.Now().Format("2006-01-02"))),
							},
							{
								Key:   aws.String("InstanceID"),
								Value: aws.String(instanceID),
							},
						},
					},
				},
			}

			_, err := s.client.CreateSnapshot(ctx, input)
			if err != nil {
				logger.Error("Failed to create snapshot", "instanceID", instanceID, "volumeID", volumeID, "error", err)
				return fmt.Errorf("failed to create snapshot for volume %s: %v", volumeID, err)
			}
			logger.Info("Created snapshot for volume", "instanceID", instanceID, "volumeID", volumeID)
		}
	}

	logger.Info("Instance backup completed successfully", "instanceID", instanceID)
	return nil
}

// InstanceConfig holds configuration for creating a new instance
type InstanceConfig struct {
	Name   string
	OSType string
	Size   string
	UserID string
}

// InstanceSummary contains information about an instance
type InstanceSummary struct {
	InstanceID   string
	Name         string
	OSType       string
	Size         string
	State        string
	LaunchTime   time.Time
	PrivateIP    string
	PublicIP     string
	CurrentAMI   string
	LatestAMI    string
	NeedsMigrate bool
}

// FormatInstanceSummary returns a human-readable string of the instance summary
func (s *InstanceSummary) FormatInstanceSummary() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Instance: %s (%s)\n", s.Name, s.InstanceID))
	b.WriteString(fmt.Sprintf("  OS:           %s\n", s.OSType))
	b.WriteString(fmt.Sprintf("  Size:         %s\n", s.Size))
	b.WriteString(fmt.Sprintf("  State:        %s\n", s.State))
	b.WriteString(fmt.Sprintf("  Launch Time:  %s\n", s.LaunchTime.Format(time.RFC3339)))
	if s.PrivateIP != "" {
		b.WriteString(fmt.Sprintf("  Private IP:   %s\n", s.PrivateIP))
	}
	if s.PublicIP != "" {
		b.WriteString(fmt.Sprintf("  Public IP:    %s\n", s.PublicIP))
	}
	b.WriteString(fmt.Sprintf("  Current AMI:  %s\n", s.CurrentAMI))
	if s.NeedsMigrate {
		b.WriteString(fmt.Sprintf("  Latest AMI:   %s (migration available)\n", s.LatestAMI))
	}

	return b.String()
}

// ListUserInstances lists all instances owned by the user
func (s *Service) ListUserInstances(ctx context.Context, userID string) ([]InstanceSummary, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:Owner"),
				Values: []string{userID},
			},
		},
	}

	result, err := s.client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("describe instances: %w", err)
	}

	var summaries []InstanceSummary
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			instanceID := aws.ToString(instance.InstanceId)

			// Get OS type
			osType, err := s.GetInstanceOSType(ctx, instanceID)
			if err != nil {
				osType = "unknown"
			}

			// Get latest AMI
			latestAMI, err := s.GetLatestAMI(ctx, osType)
			if err != nil {
				latestAMI = "unknown"
			}

			// Get instance name from tags
			name := instanceID
			for _, tag := range instance.Tags {
				if aws.ToString(tag.Key) == "Name" {
					name = aws.ToString(tag.Value)
					break
				}
			}

			summary := InstanceSummary{
				InstanceID:   instanceID,
				Name:         name,
				OSType:       osType,
				Size:         string(instance.InstanceType),
				State:        string(instance.State.Name),
				LaunchTime:   aws.ToTime(instance.LaunchTime),
				PrivateIP:    aws.ToString(instance.PrivateIpAddress),
				PublicIP:     aws.ToString(instance.PublicIpAddress),
				CurrentAMI:   aws.ToString(instance.ImageId),
				LatestAMI:    latestAMI,
				NeedsMigrate: aws.ToString(instance.ImageId) != latestAMI,
			}
			summaries = append(summaries, summary)
		}
	}

	return summaries, nil
}

// CreateInstance creates a new EC2 instance with the given configuration
func (s *Service) CreateInstance(ctx context.Context, config InstanceConfig) (*InstanceSummary, error) {
	// Validate size
	if !isValidSize(config.Size) {
		return nil, fmt.Errorf("invalid size: %s. Must be one of: small, medium, large, xlarge", config.Size)
	}

	// Get the latest AMI for the OS type
	amiInfo, err := s.GetAMIWithTag(ctx, "OS", config.OSType)
	if err != nil {
		return nil, fmt.Errorf("get latest AMI: %w", err)
	}

	// Get instance type based on size
	instanceType := getInstanceType(config.Size)

	// Create the instance
	input := &ec2.RunInstancesInput{
		ImageId:      amiInfo.ImageId,
		InstanceType: types.InstanceType(instanceType),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(config.Name),
					},
					{
						Key:   aws.String("Owner"),
						Value: aws.String(config.UserID),
					},
				},
			},
		},
	}

	output, err := s.client.RunInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("run instance: %w", err)
	}

	if len(output.Instances) == 0 {
		return nil, fmt.Errorf("no instance created")
	}

	instance := output.Instances[0]
	amiID := aws.ToString(amiInfo.ImageId)
	return &InstanceSummary{
		InstanceID:   aws.ToString(instance.InstanceId),
		Name:         config.Name,
		OSType:       config.OSType,
		Size:         config.Size,
		State:        string(instance.State.Name),
		LaunchTime:   aws.ToTime(instance.LaunchTime),
		PrivateIP:    aws.ToString(instance.PrivateIpAddress),
		PublicIP:     aws.ToString(instance.PublicIpAddress),
		CurrentAMI:   amiID,
		LatestAMI:    amiID,
		NeedsMigrate: false,
	}, nil
}

func isValidSize(size string) bool {
	validSizes := []string{"small", "medium", "large", "xlarge"}
	for _, s := range validSizes {
		if s == size {
			return true
		}
	}
	return false
}

func getInstanceType(size string) string {
	switch size {
	case "small":
		return "t2.micro"
	case "medium":
		return "t2.small"
	case "large":
		return "t2.medium"
	case "xlarge":
		return "t2.large"
	default:
		return "t2.micro"
	}
}

// CheckMigrationStatus checks if a user's instance needs migration
func (s *Service) CheckMigrationStatus(ctx context.Context, userID string) (*MigrationStatus, error) {
	// Find user's instance
	input := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:Owner"),
				Values: []string{userID},
			},
		},
	}

	result, err := s.client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("describe instances: %w", err)
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("no instance found for user: %s", userID)
	}

	instance := result.Reservations[0].Instances[0]
	instanceID := aws.ToString(instance.InstanceId)

	// Get instance OS type
	osType, err := s.GetInstanceOSType(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("get OS type: %w", err)
	}

	// Get latest AMI for this OS
	latestAMI, err := s.GetLatestAMI(ctx, osType)
	if err != nil {
		return nil, fmt.Errorf("get latest AMI: %w", err)
	}

	// Get latest AMI details
	latestAMIDetails, err := s.getAMIDetails(ctx, latestAMI)
	if err != nil {
		return nil, fmt.Errorf("get AMI details: %w", err)
	}

	// Get current AMI details
	currentAMI := aws.ToString(instance.ImageId)
	currentAMIDetails, err := s.getAMIDetails(ctx, currentAMI)
	if err != nil {
		return nil, fmt.Errorf("get current AMI details: %w", err)
	}

	status := &MigrationStatus{
		InstanceID:     instanceID,
		OSType:         osType,
		CurrentAMI:     currentAMI,
		LatestAMI:      latestAMI,
		NeedsMigration: currentAMI != latestAMI,
		CurrentAMIInfo: currentAMIDetails,
		LatestAMIInfo:  latestAMIDetails,
		InstanceState:  string(instance.State.Name),
		InstanceType:   string(instance.InstanceType),
		LaunchTime:     aws.ToTime(instance.LaunchTime),
		PrivateIP:      aws.ToString(instance.PrivateIpAddress),
		PublicIP:       aws.ToString(instance.PublicIpAddress),
	}

	return status, nil
}

// MigrationStatus contains information about an instance's migration status
type MigrationStatus struct {
	InstanceID     string
	OSType         string
	CurrentAMI     string
	LatestAMI      string
	NeedsMigration bool
	CurrentAMIInfo *AMIDetails
	LatestAMIInfo  *AMIDetails
	InstanceState  string
	InstanceType   string
	LaunchTime     time.Time
	PrivateIP      string
	PublicIP       string
}

// AMIDetails contains information about an AMI
type AMIDetails struct {
	Name        string
	Description string
	CreatedDate string
	Tags        map[string]string
}

func (s *Service) getAMIDetails(ctx context.Context, amiID string) (*AMIDetails, error) {
	input := &ec2.DescribeImagesInput{
		ImageIds: []string{amiID},
	}

	result, err := s.client.DescribeImages(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("describe images: %w", err)
	}

	if len(result.Images) == 0 {
		return nil, fmt.Errorf("AMI not found: %s", amiID)
	}

	image := result.Images[0]
	tags := make(map[string]string)
	for _, tag := range image.Tags {
		tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}

	return &AMIDetails{
		Name:        aws.ToString(image.Name),
		Description: aws.ToString(image.Description),
		CreatedDate: aws.ToString(image.CreationDate),
		Tags:        tags,
	}, nil
}

// FormatMigrationStatus returns a human-readable string of the migration status
func (s *MigrationStatus) FormatMigrationStatus() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Instance Status for %s:\n", s.InstanceID))
	b.WriteString(fmt.Sprintf("  OS Type:        %s\n", s.OSType))
	b.WriteString(fmt.Sprintf("  Instance Type:  %s\n", s.InstanceType))
	b.WriteString(fmt.Sprintf("  State:          %s\n", s.InstanceState))
	b.WriteString(fmt.Sprintf("  Launch Time:    %s\n", s.LaunchTime.Format(time.RFC3339)))
	if s.PrivateIP != "" {
		b.WriteString(fmt.Sprintf("  Private IP:     %s\n", s.PrivateIP))
	}
	if s.PublicIP != "" {
		b.WriteString(fmt.Sprintf("  Public IP:      %s\n", s.PublicIP))
	}
	b.WriteString("\nAMI Status:\n")
	b.WriteString(fmt.Sprintf("  Current AMI:    %s\n", s.CurrentAMI))
	if s.CurrentAMIInfo != nil {
		b.WriteString(fmt.Sprintf("    Name:         %s\n", s.CurrentAMIInfo.Name))
		b.WriteString(fmt.Sprintf("    Created:      %s\n", s.CurrentAMIInfo.CreatedDate))
	}
	b.WriteString(fmt.Sprintf("  Latest AMI:     %s\n", s.LatestAMI))
	if s.LatestAMIInfo != nil {
		b.WriteString(fmt.Sprintf("    Name:         %s\n", s.LatestAMIInfo.Name))
		b.WriteString(fmt.Sprintf("    Created:      %s\n", s.LatestAMIInfo.CreatedDate))
	}
	b.WriteString(fmt.Sprintf("\nMigration Needed: %v\n", s.NeedsMigration))
	if s.NeedsMigration {
		b.WriteString("\nRun 'ami-migrate migrate' to update your instance to the latest AMI.")
	}

	return b.String()
}

// DeleteInstance deletes an instance owned by the user
func (s *Service) DeleteInstance(ctx context.Context, userID, instanceID string) error {
	// First verify the instance belongs to the user
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:Owner"),
				Values: []string{userID},
			},
		},
	}

	result, err := s.client.DescribeInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("describe instance: %w", err)
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return fmt.Errorf("instance %s not found or not owned by user %s", instanceID, userID)
	}

	instance := result.Reservations[0].Instances[0]

	// Check if instance is already terminated
	if instance.State != nil && instance.State.Name == types.InstanceStateNameTerminated {
		return fmt.Errorf("instance %s is already terminated", instanceID)
	}

	// Terminate the instance
	terminateInput := &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	}

	_, err = s.client.TerminateInstances(ctx, terminateInput)
	if err != nil {
		return fmt.Errorf("terminate instance: %w", err)
	}

	return nil
}

func hasTag(tags []types.Tag, key, value string) bool {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key && aws.ToString(tag.Value) == value {
			return true
		}
	}
	return false
}

type waiterInterface interface {
	Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration) error
}

type runningWaiter struct {
	*ec2.InstanceRunningWaiter
}

func (w *runningWaiter) Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration) error {
	return w.InstanceRunningWaiter.Wait(ctx, params, maxWaitDur)
}

type stoppedWaiter struct {
	*ec2.InstanceStoppedWaiter
}

func (w *stoppedWaiter) Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration) error {
	return w.InstanceStoppedWaiter.Wait(ctx, params, maxWaitDur)
}

type terminatedWaiter struct {
	*ec2.InstanceTerminatedWaiter
}

func (w *terminatedWaiter) Wait(ctx context.Context, params *ec2.DescribeInstancesInput, maxWaitDur time.Duration) error {
	return w.InstanceTerminatedWaiter.Wait(ctx, params, maxWaitDur)
}

// waitForInstanceState waits for an instance to reach the desired state
func waitForInstanceState(ctx context.Context, client ecTypes.EC2ClientAPI, instanceID string, desiredState types.InstanceStateName) error {
	// For mock client, return immediately since state transitions are immediate
	if _, ok := client.(*ecTypes.MockEC2Client); ok {
		return nil
	}

	var waiter waiterInterface

	switch desiredState {
	case types.InstanceStateNameRunning:
		waiter = &runningWaiter{ec2.NewInstanceRunningWaiter(client)}
	case types.InstanceStateNameStopped:
		waiter = &stoppedWaiter{ec2.NewInstanceStoppedWaiter(client)}
	case types.InstanceStateNameTerminated:
		waiter = &terminatedWaiter{ec2.NewInstanceTerminatedWaiter(client)}
	default:
		return fmt.Errorf("unsupported instance state: %s", desiredState)
	}

	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	if err := waiter.Wait(ctx, input, config.GetTimeout()); err != nil {
		return fmt.Errorf("failed to wait for instance %s to reach state %s: %w", instanceID, desiredState, err)
	}

	return nil
}
