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
)

// EC2ClientAPI defines the AWS EC2 client interface
type EC2ClientAPI interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
	CreateSnapshot(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error)
	StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
	StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error)
	RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
	TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
	DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error)
	CreateVolume(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error)
	DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error)
	AttachVolume(ctx context.Context, params *ec2.AttachVolumeInput, optFns ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error)
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
func (s *Service) MigrateInstances(ctx context.Context, enabledValue string) error {
	instances, err := s.fetchEnabledInstances(ctx, enabledValue)
	if err != nil {
		return fmt.Errorf("fetch instances: %w", err)
	}

	if len(instances) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(instances))

	for _, instance := range instances {
		wg.Add(1)
		go func(inst types.Instance) {
			defer wg.Done()
			instanceID := aws.ToString(inst.InstanceId)
			if err := s.MigrateInstance(ctx, instanceID); err != nil {
				errChan <- fmt.Errorf("migrate instance %s: %w", instanceID, err)
			}
		}(instance)
	}

	wg.Wait()
	close(errChan)

	// Collect any errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("migration errors: %v", errs)
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
	}, 5*time.Minute); err != nil {
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
		}, 5*time.Minute); err != nil {
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

func hasTag(tags []types.Tag, key, value string) bool {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key && aws.ToString(tag.Value) == value {
			return true
		}
	}
	return false
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

func (s *Service) MigrateInstance(ctx context.Context, instanceID string) error {
	// Determine the OS type
	osType, err := s.GetInstanceOSType(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("get OS type: %w", err)
	}

	// Get the latest AMI for this OS type
	latestAMI, err := s.GetLatestAMI(ctx, osType)
	if err != nil {
		return fmt.Errorf("get latest AMI: %w", err)
	}

	// Get the current AMI ID
	instance, err := s.getInstance(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("get instance: %w", err)
	}

	currentAMI := aws.ToString(instance.ImageId)
	if currentAMI == latestAMI {
		return nil // Already on latest AMI
	}

	// Perform the migration
	return s.migrateInstanceToAMI(ctx, instance, latestAMI)
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
	instanceID := aws.ToString(instance.InstanceId)
	
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
	return nil
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
		InstanceID:        instanceID,
		OSType:           osType,
		CurrentAMI:       currentAMI,
		LatestAMI:        latestAMI,
		NeedsMigration:   currentAMI != latestAMI,
		CurrentAMIInfo:   currentAMIDetails,
		LatestAMIInfo:    latestAMIDetails,
		InstanceState:    string(instance.State.Name),
		InstanceType:     aws.ToString(instance.InstanceType),
		LaunchTime:       aws.ToTime(instance.LaunchTime),
		PrivateIP:        aws.ToString(instance.PrivateIpAddress),
		PublicIP:         aws.ToString(instance.PublicIpAddress),
	}

	return status, nil
}

// MigrationStatus contains information about an instance's migration status
type MigrationStatus struct {
	InstanceID      string
	OSType          string
	CurrentAMI      string
	LatestAMI       string
	NeedsMigration  bool
	CurrentAMIInfo  *AMIDetails
	LatestAMIInfo   *AMIDetails
	InstanceState   string
	InstanceType    string
	LaunchTime      time.Time
	PrivateIP       string
	PublicIP        string
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
