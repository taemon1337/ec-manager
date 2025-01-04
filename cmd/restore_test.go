package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	mockclient "github.com/taemon1337/ec-manager/pkg/mock"
	"github.com/taemon1337/ec-manager/pkg/mock/fixtures"
	"github.com/taemon1337/ec-manager/pkg/mock/waiters"
	"github.com/taemon1337/ec-manager/pkg/testutil"
)

func NewRestoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore an EC2 instance from a snapshot",
		Long:  "Create a new EC2 instance from a snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			instanceID, err := cmd.Flags().GetString("instance-id")
			if err != nil {
				return err
			}

			if instanceID == "" {
				return fmt.Errorf("instance ID must be set")
			}

			amiID, err := cmd.Flags().GetString("ami")
			if err != nil {
				return err
			}

			snapshotID, err := cmd.Flags().GetString("snapshot")
			if err != nil {
				return err
			}

			// If neither AMI nor snapshot is provided, return error
			if amiID == "" && snapshotID == "" {
				return fmt.Errorf("either AMI ID or snapshot ID must be set")
			}

			mockClient, ok := ctx.Value(mockclient.EC2ClientKey).(*mockclient.MockEC2Client)
			if !ok {
				return fmt.Errorf("failed to get EC2 client")
			}

			// Get instance details
			output, err := mockClient.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
				InstanceIds: []string{instanceID},
			}, func(options *ec2.Options) {})
			if err != nil {
				return err
			}

			if len(output.Reservations) == 0 || len(output.Reservations[0].Instances) == 0 {
				return fmt.Errorf("instance not found")
			}

			if snapshotID != "" {
				// Stop the instance
				_, err = mockClient.StopInstances(ctx, &ec2.StopInstancesInput{
					InstanceIds: []string{instanceID},
				}, func(options *ec2.Options) {})
				if err != nil {
					return fmt.Errorf("failed to stop instance: %v", err)
				}

				// Wait for instance to stop
				waiter := mockClient.NewInstanceStoppedWaiter()
				err = waiter.Wait(ctx, &ec2.DescribeInstancesInput{
					InstanceIds: []string{instanceID},
				}, 0, func(options *ec2.InstanceStoppedWaiterOptions) {})
				if err != nil {
					return fmt.Errorf("failed waiting for instance to stop: %v", err)
				}

				// Get snapshot details
				var snapshotOutput *ec2.DescribeSnapshotsOutput
				snapshotOutput, err = mockClient.DescribeSnapshots(ctx, &ec2.DescribeSnapshotsInput{
					SnapshotIds: []string{snapshotID},
				}, func(options *ec2.Options) {})
				if err != nil {
					return err
				}

				if len(snapshotOutput.Snapshots) == 0 {
					return fmt.Errorf("snapshot not found")
				}

				// Create new volume from snapshot
				var volumeOutput *ec2.CreateVolumeOutput
				volumeOutput, err = mockClient.CreateVolume(ctx, &ec2.CreateVolumeInput{
					SnapshotId:       aws.String(snapshotID),
					AvailabilityZone: output.Reservations[0].Instances[0].Placement.AvailabilityZone,
				}, func(options *ec2.Options) {})
				if err != nil {
					return err
				}

				// Wait for volume to become available
				volumeWaiter := mockClient.NewVolumeAvailableWaiter()
				err = volumeWaiter.Wait(ctx, &ec2.DescribeVolumesInput{
					VolumeIds: []string{*volumeOutput.VolumeId},
				}, 0, func(options *ec2.VolumeAvailableWaiterOptions) {})
				if err != nil {
					return fmt.Errorf("failed waiting for volume to become available: %v", err)
				}

				// Attach volume to instance
				_, err = mockClient.AttachVolume(ctx, &ec2.AttachVolumeInput{
					InstanceId: output.Reservations[0].Instances[0].InstanceId,
					VolumeId:   volumeOutput.VolumeId,
					Device:     snapshotOutput.Snapshots[0].Tags[0].Value,
				}, func(options *ec2.Options) {})
				if err != nil {
					return err
				}

				// Start the instance
				_, err = mockClient.StartInstances(ctx, &ec2.StartInstancesInput{
					InstanceIds: []string{instanceID},
				}, func(options *ec2.Options) {})
				if err != nil {
					return fmt.Errorf("failed to start instance: %v", err)
				}

				// Wait for instance to start
				runningWaiter := mockClient.NewInstanceRunningWaiter()
				err = runningWaiter.Wait(ctx, &ec2.DescribeInstancesInput{
					InstanceIds: []string{instanceID},
				}, 0, func(options *ec2.InstanceRunningWaiterOptions) {})
				if err != nil {
					return fmt.Errorf("failed waiting for instance to start: %v", err)
				}

				return nil
			}

			// If using AMI
			// Get AMI details
			amiOutput, err := mockClient.DescribeImages(ctx, &ec2.DescribeImagesInput{
				ImageIds: []string{amiID},
			}, func(options *ec2.Options) {})
			if err != nil {
				return fmt.Errorf("failed to describe AMI: %v", err)
			}

			if len(amiOutput.Images) == 0 {
				return fmt.Errorf("AMI not found")
			}

			// Create new instance
			newInstance, err := mockClient.RunInstances(ctx, &ec2.RunInstancesInput{
				ImageId:      aws.String(amiID),
				InstanceType: output.Reservations[0].Instances[0].InstanceType,
				MinCount:     aws.Int32(1),
				MaxCount:     aws.Int32(1),
				SubnetId:     output.Reservations[0].Instances[0].SubnetId,
			}, func(options *ec2.Options) {})
			if err != nil {
				return err
			}

			// Create tags for new instance
			_, err = mockClient.CreateTags(ctx, &ec2.CreateTagsInput{
				Resources: []string{*newInstance.Instances[0].InstanceId},
				Tags:      output.Reservations[0].Instances[0].Tags,
			}, nil)
			if err != nil {
				return err
			}

			// Stop original instance
			_, err = mockClient.StopInstances(ctx, &ec2.StopInstancesInput{
				InstanceIds: []string{instanceID},
			}, func(options *ec2.Options) {})
			if err != nil {
				return err
			}

			// Start new instance
			_, err = mockClient.StartInstances(ctx, &ec2.StartInstancesInput{
				InstanceIds: []string{*newInstance.Instances[0].InstanceId},
			}, func(options *ec2.Options) {})
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringP("instance-id", "i", "", "Instance to restore")
	cmd.Flags().StringP("ami", "a", "", "AMI to restore from")
	cmd.Flags().StringP("snapshot", "s", "", "Snapshot to restore from")
	return cmd
}

func TestRestoreCmd(t *testing.T) {
	tests := []testutil.CommandTestCase{
		{
			Name: "success",
			Args: []string{"--instance-id", "i-123", "--ami", "ami-123"},
			MockEC2Setup: func(m *mockclient.MockEC2Client) {
				// Mock instance lookup
				m.On("DescribeInstances", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{fixtures.TestInstance()},
						},
					},
				}, nil).Once()

				// Mock AMI lookup
				m.On("DescribeImages", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeImagesInput) bool {
					return len(input.ImageIds) == 1 && input.ImageIds[0] == "ami-123"
				}), mock.Anything).Return(&ec2.DescribeImagesOutput{
					Images: []types.Image{fixtures.TestAMI()},
				}, nil).Once()

				// Mock RunInstances
				m.On("RunInstances", mock.Anything, mock.MatchedBy(func(input *ec2.RunInstancesInput) bool {
					return *input.ImageId == "ami-123" && input.InstanceType == types.InstanceTypeT2Micro
				}), mock.Anything).Return(&ec2.RunInstancesOutput{
					Instances: []types.Instance{
						{
							InstanceId: aws.String("i-mock123"),
						},
					},
				}, nil).Once()

				// Mock CreateTags
				m.On("CreateTags", mock.Anything, mock.MatchedBy(func(input *ec2.CreateTagsInput) bool {
					return len(input.Resources) == 1 && input.Resources[0] == "i-mock123" && len(input.Tags) == len(fixtures.TestInstance().Tags)
				}), mock.Anything).Return(&ec2.CreateTagsOutput{}, nil).Once()

				// Mock StopInstances
				m.On("StopInstances", mock.Anything, mock.MatchedBy(func(input *ec2.StopInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.Anything).Return(&ec2.StopInstancesOutput{}, nil).Once()

				// Mock StartInstances
				m.On("StartInstances", mock.Anything, mock.MatchedBy(func(input *ec2.StartInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-mock123"
				}), mock.Anything).Return(&ec2.StartInstancesOutput{}, nil).Once()
			},
			SetupContext: func(ctx context.Context) context.Context {
				mockClient := &mockclient.MockEC2Client{}

				// Mock instance stopped waiter
				stoppedWaiter := &waiters.MockInstanceStoppedWaiter{}
				stoppedWaiter.On("Wait", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.AnythingOfType("time.Duration"), mock.AnythingOfType("[]func(*ec2.InstanceStoppedWaiterOptions)")).Return(nil)
				mockClient.InstanceStoppedWaiter = stoppedWaiter

				// Mock instance running waiter
				runningWaiter := &waiters.MockInstanceRunningWaiter{}
				runningWaiter.On("Wait", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.AnythingOfType("time.Duration"), mock.AnythingOfType("[]func(*ec2.InstanceRunningWaiterOptions)")).Return(nil)
				mockClient.InstanceRunningWaiter = runningWaiter

				return context.WithValue(ctx, mockclient.EC2ClientKey, mockClient)
			},
		},
		{
			Name:        "instance_not_found",
			Args:        []string{"--instance-id", "i-nonexistent", "--ami", "ami-123"},
			WantErr:     true,
			ErrContains: "instance not found",
			MockEC2Setup: func(m *mockclient.MockEC2Client) {
				// Mock instance lookup
				m.On("DescribeInstances", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-nonexistent"
				}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
				}, nil).Once()
			},
			SetupContext: func(ctx context.Context) context.Context {
				mockClient := &mockclient.MockEC2Client{}

				// Mock instance stopped waiter
				stoppedWaiter := &waiters.MockInstanceStoppedWaiter{}
				stoppedWaiter.On("Wait", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-nonexistent"
				}), mock.AnythingOfType("time.Duration"), mock.AnythingOfType("[]func(*ec2.InstanceStoppedWaiterOptions)")).Return(nil)
				mockClient.InstanceStoppedWaiter = stoppedWaiter

				// Mock instance running waiter
				runningWaiter := &waiters.MockInstanceRunningWaiter{}
				runningWaiter.On("Wait", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-nonexistent"
				}), mock.AnythingOfType("time.Duration"), mock.AnythingOfType("[]func(*ec2.InstanceRunningWaiterOptions)")).Return(nil)
				mockClient.InstanceRunningWaiter = runningWaiter

				return context.WithValue(ctx, mockclient.EC2ClientKey, mockClient)
			},
		},
		{
			Name:        "ami_not_found",
			Args:        []string{"--instance-id", "i-123", "--ami", "ami-nonexistent"},
			WantErr:     true,
			ErrContains: "AMI not found",
			MockEC2Setup: func(m *mockclient.MockEC2Client) {
				// Mock instance lookup
				m.On("DescribeInstances", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{fixtures.TestInstance()},
						},
					},
				}, nil).Once()

				// Mock AMI lookup
				m.On("DescribeImages", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeImagesInput) bool {
					return len(input.ImageIds) == 1 && input.ImageIds[0] == "ami-nonexistent"
				}), mock.Anything).Return(&ec2.DescribeImagesOutput{
					Images: []types.Image{},
				}, nil).Once()
			},
			SetupContext: func(ctx context.Context) context.Context {
				mockClient := &mockclient.MockEC2Client{}

				// Mock instance stopped waiter
				stoppedWaiter := &waiters.MockInstanceStoppedWaiter{}
				stoppedWaiter.On("Wait", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.AnythingOfType("time.Duration"), mock.AnythingOfType("[]func(*ec2.InstanceStoppedWaiterOptions)")).Return(nil)
				mockClient.InstanceStoppedWaiter = stoppedWaiter

				// Mock instance running waiter
				runningWaiter := &waiters.MockInstanceRunningWaiter{}
				runningWaiter.On("Wait", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.AnythingOfType("time.Duration"), mock.AnythingOfType("[]func(*ec2.InstanceRunningWaiterOptions)")).Return(nil)
				mockClient.InstanceRunningWaiter = runningWaiter

				return context.WithValue(ctx, mockclient.EC2ClientKey, mockClient)
			},
		},
		{
			Name: "restore_from_snapshot_success",
			Args: []string{"--instance-id", "i-123", "--snapshot", "snap-123"},
			MockEC2Setup: func(m *mockclient.MockEC2Client) {
				// Mock instance lookup
				m.On("DescribeInstances", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{fixtures.TestInstance()},
						},
					},
				}, nil).Once()

				// Mock StopInstances
				m.On("StopInstances", mock.Anything, mock.MatchedBy(func(input *ec2.StopInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.Anything).Return(&ec2.StopInstancesOutput{}, nil).Once()

				// Mock DescribeSnapshots
				snapshot := fixtures.TestSnapshot()
				// Move the device tag to be first in the list
				snapshot.Tags = []types.Tag{
					{
						Key:   aws.String("ami-migrate-device"),
						Value: aws.String("/dev/xvdf"),
					},
					{
						Key:   aws.String("Name"),
						Value: aws.String("test-snapshot"),
					},
				}
				m.On("DescribeSnapshots", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeSnapshotsInput) bool {
					return len(input.SnapshotIds) == 1 && input.SnapshotIds[0] == "snap-123"
				}), mock.Anything).Return(&ec2.DescribeSnapshotsOutput{
					Snapshots: []types.Snapshot{snapshot},
				}, nil).Once()

				// Mock CreateVolume
				m.On("CreateVolume", mock.Anything, mock.MatchedBy(func(input *ec2.CreateVolumeInput) bool {
					return *input.SnapshotId == "snap-123" && *input.AvailabilityZone == "us-east-1a"
				}), mock.Anything).Return(&ec2.CreateVolumeOutput{
					VolumeId: aws.String("vol-new123"),
				}, nil).Once()

				// Mock AttachVolume
				m.On("AttachVolume", mock.Anything, mock.MatchedBy(func(input *ec2.AttachVolumeInput) bool {
					return *input.InstanceId == "i-123" && *input.VolumeId == "vol-new123" && *input.Device == "/dev/xvdf"
				}), mock.Anything).Return(&ec2.AttachVolumeOutput{}, nil).Once()

				// Mock StartInstances
				m.On("StartInstances", mock.Anything, mock.MatchedBy(func(input *ec2.StartInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.Anything).Return(&ec2.StartInstancesOutput{}, nil).Once()

				// Setup waiters
				stoppedWaiter := &waiters.MockInstanceStoppedWaiter{}
				stoppedWaiter.On("Wait", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.AnythingOfType("time.Duration"), mock.AnythingOfType("[]func(*ec2.InstanceStoppedWaiterOptions)")).Return(nil)
				m.InstanceStoppedWaiter = stoppedWaiter

				volumeWaiter := &waiters.MockVolumeAvailableWaiter{}
				volumeWaiter.On("Wait", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeVolumesInput) bool {
					return len(input.VolumeIds) == 1 && input.VolumeIds[0] == "vol-new123"
				}), mock.AnythingOfType("time.Duration"), mock.AnythingOfType("[]func(*ec2.VolumeAvailableWaiterOptions)")).Return(nil)
				m.VolumeAvailableWaiter = volumeWaiter

				runningWaiter := &waiters.MockInstanceRunningWaiter{}
				runningWaiter.On("Wait", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.AnythingOfType("time.Duration"), mock.AnythingOfType("[]func(*ec2.InstanceRunningWaiterOptions)")).Return(nil)
				m.InstanceRunningWaiter = runningWaiter
			},
			SetupContext: func(ctx context.Context) context.Context {
				mockClient := &mockclient.MockEC2Client{}
				return context.WithValue(ctx, mockclient.EC2ClientKey, mockClient)
			},
		},
		{
			Name:        "restore_from_snapshot_stop_error",
			Args:        []string{"--instance-id", "i-123", "--snapshot", "snap-123"},
			WantErr:     true,
			ErrContains: "failed to stop instance",
			MockEC2Setup: func(m *mockclient.MockEC2Client) {
				// Mock instance lookup
				m.On("DescribeInstances", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{fixtures.TestInstance()},
						},
					},
				}, nil).Once()

				// Mock StopInstances failure
				m.On("StopInstances", mock.Anything, mock.MatchedBy(func(input *ec2.StopInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.Anything).Return(&ec2.StopInstancesOutput{}, fmt.Errorf("failed to stop instance")).Once()
			},
			SetupContext: func(ctx context.Context) context.Context {
				mockClient := &mockclient.MockEC2Client{}

				// Mock instance stopped waiter
				stoppedWaiter := &waiters.MockInstanceStoppedWaiter{}
				stoppedWaiter.On("Wait", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.AnythingOfType("time.Duration"), mock.AnythingOfType("[]func(*ec2.InstanceStoppedWaiterOptions)")).Return(nil)
				mockClient.InstanceStoppedWaiter = stoppedWaiter

				// Mock instance running waiter
				runningWaiter := &waiters.MockInstanceRunningWaiter{}
				runningWaiter.On("Wait", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.AnythingOfType("time.Duration"), mock.AnythingOfType("[]func(*ec2.InstanceRunningWaiterOptions)")).Return(nil)
				mockClient.InstanceRunningWaiter = runningWaiter

				return context.WithValue(ctx, mockclient.EC2ClientKey, mockClient)
			},
		},
		{
			Name: "success_with_snapshot",
			Args: []string{"--instance-id", "i-123", "--snapshot", "snap-123"},
			MockEC2Setup: func(m *mockclient.MockEC2Client) {
				// Mock instance lookup
				m.On("DescribeInstances", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{fixtures.TestInstance()},
						},
					},
				}, nil).Once()

				// Mock StopInstances
				m.On("StopInstances", mock.Anything, mock.MatchedBy(func(input *ec2.StopInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.Anything).Return(&ec2.StopInstancesOutput{}, nil).Once()

				// Mock DescribeSnapshots
				m.On("DescribeSnapshots", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeSnapshotsInput) bool {
					return len(input.SnapshotIds) == 1 && input.SnapshotIds[0] == "snap-123"
				}), mock.Anything).Return(&ec2.DescribeSnapshotsOutput{
					Snapshots: []types.Snapshot{fixtures.TestSnapshot()},
				}, nil).Once()

				// Mock CreateVolume
				m.On("CreateVolume", mock.Anything, mock.MatchedBy(func(input *ec2.CreateVolumeInput) bool {
					return *input.SnapshotId == "snap-123" && *input.AvailabilityZone == "us-east-1a"
				}), mock.Anything).Return(&ec2.CreateVolumeOutput{
					VolumeId: aws.String("vol-new123"),
				}, nil).Once()

				// Mock AttachVolume
				m.On("AttachVolume", mock.Anything, mock.MatchedBy(func(input *ec2.AttachVolumeInput) bool {
					return *input.InstanceId == "i-123" && *input.VolumeId == "vol-new123"
				}), mock.Anything).Return(&ec2.AttachVolumeOutput{}, nil).Once()

				// Mock StartInstances
				m.On("StartInstances", mock.Anything, mock.MatchedBy(func(input *ec2.StartInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.Anything).Return(&ec2.StartInstancesOutput{}, nil).Once()
			},
			SetupContext: func(ctx context.Context) context.Context {
				mockClient := &mockclient.MockEC2Client{}

				// Mock instance stopped waiter
				stoppedWaiter := &waiters.MockInstanceStoppedWaiter{}
				stoppedWaiter.On("Wait", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.AnythingOfType("time.Duration"), mock.AnythingOfType("[]func(*ec2.InstanceStoppedWaiterOptions)")).Return(nil)
				mockClient.InstanceStoppedWaiter = stoppedWaiter

				// Mock volume available waiter
				volumeWaiter := &waiters.MockVolumeAvailableWaiter{}
				volumeWaiter.On("Wait", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeVolumesInput) bool {
					return len(input.VolumeIds) == 1 && input.VolumeIds[0] == "vol-new123"
				}), mock.AnythingOfType("time.Duration"), mock.AnythingOfType("[]func(*ec2.VolumeAvailableWaiterOptions)")).Return(nil)
				mockClient.VolumeAvailableWaiter = volumeWaiter

				// Mock instance running waiter
				runningWaiter := &waiters.MockInstanceRunningWaiter{}
				runningWaiter.On("Wait", mock.Anything, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
					return len(input.InstanceIds) == 1 && input.InstanceIds[0] == "i-123"
				}), mock.AnythingOfType("time.Duration"), mock.AnythingOfType("[]func(*ec2.InstanceRunningWaiterOptions)")).Return(nil)
				mockClient.InstanceRunningWaiter = runningWaiter

				return context.WithValue(ctx, mockclient.EC2ClientKey, mockClient)
			},
		},
	}

	testutil.RunCommandTest(t, NewRestoreCmd, tests)
}
