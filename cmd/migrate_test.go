package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/taemon1337/ami-migrate/pkg/ami"
	"github.com/taemon1337/ami-migrate/pkg/client"
	"github.com/taemon1337/ami-migrate/pkg/logger"
	"github.com/taemon1337/ami-migrate/pkg/testutil"
	apitypes "github.com/taemon1337/ami-migrate/pkg/types"
)

func TestMigrateCmd(t *testing.T) {
	// Initialize test logger
	testutil.InitTestLogger(t)

	// Store original values
	originalInstanceID := instanceID
	originalEnabled := enabled
	originalLogLevel := logLevel
	originalNewAMI := newAMI

	// Reset flags after tests
	t.Cleanup(func() {
		instanceID = originalInstanceID
		enabled = originalEnabled
		logLevel = originalLogLevel
		newAMI = originalNewAMI
		client.ResetClient()
	})

	// Initialize logger for tests with debug level
	logLevel = "debug"
	logger.Init(logger.LogLevel(logLevel))

	tests := []struct {
		name      string
		setupCmd  func(*cobra.Command)
		setupMock func(*apitypes.MockEC2Client)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful migration",
			setupCmd: func(cmd *cobra.Command) {
				cmd.Flags().Set("instance-id", "i-123")
				cmd.Flags().Set("new-ami", "ami-456")
			},
			setupMock: func(m *apitypes.MockEC2Client) {
				instance := types.Instance{
					InstanceId:   aws.String("i-123"),
					InstanceType: types.InstanceTypeT2Micro,
					ImageId:      aws.String("ami-123"), // Current AMI
					State: &types.InstanceState{
						Name: types.InstanceStateNameRunning,
					},
					BlockDeviceMappings: []types.InstanceBlockDeviceMapping{
						{
							DeviceName: aws.String("/dev/xvda"),
							Ebs: &types.EbsInstanceBlockDevice{
								VolumeId: aws.String("vol-123"),
							},
						},
					},
					Tags: []types.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String("test-instance"),
						},
					},
				}
				m.Instances = []types.Instance{instance}
				m.SetInstanceState("i-123", types.InstanceStateNameRunning)

				// Set up DescribeInstances response
				m.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{instance},
						},
					},
				}

				// Set up RunInstances response
				newInstance := instance
				newInstance.InstanceId = aws.String("i-456")
				newInstance.ImageId = aws.String("ami-456")
				m.RunInstancesOutput = &ec2.RunInstancesOutput{
					Instances: []types.Instance{newInstance},
				}

				// Set up TerminateInstances response
				m.TerminateInstancesOutput = &ec2.TerminateInstancesOutput{
					TerminatingInstances: []types.InstanceStateChange{
						{
							CurrentState: &types.InstanceState{
								Name: types.InstanceStateNameShuttingDown,
							},
							InstanceId: aws.String("i-123"),
							PreviousState: &types.InstanceState{
								Name: types.InstanceStateNameRunning,
							},
						},
					},
				}

				// Set up CreateSnapshot response
				m.CreateSnapshotOutput = &ec2.CreateSnapshotOutput{
					SnapshotId: aws.String("snap-123"),
				}

				// Set up CreateTags response
				m.CreateTagsOutput = &ec2.CreateTagsOutput{}

				// Set up StopInstances response
				m.StopInstancesOutput = &ec2.StopInstancesOutput{
					StoppingInstances: []types.InstanceStateChange{
						{
							CurrentState: &types.InstanceState{
								Name: types.InstanceStateNameStopping,
							},
							InstanceId: aws.String("i-123"),
							PreviousState: &types.InstanceState{
								Name: types.InstanceStateNameRunning,
							},
						},
					},
				}
			},
		},
		{
			name: "missing required flags",
			setupCmd: func(cmd *cobra.Command) {
				// Don't set any flags
			},
			wantErr: true,
			errMsg:  "either --instance-id or --enabled flag must be specified",
		},
		{
			name: "instance not found",
			setupCmd: func(cmd *cobra.Command) {
				cmd.Flags().Set("instance-id", "i-123")
				cmd.Flags().Set("new-ami", "ami-456")
			},
			setupMock: func(m *apitypes.MockEC2Client) {
				m.Instances = []types.Instance{} // Empty instance list
			},
			wantErr: true,
			errMsg:  "instance not found",
		},
		{
			name: "error stopping instance",
			setupCmd: func(cmd *cobra.Command) {
				cmd.Flags().Set("instance-id", "i-123")
				cmd.Flags().Set("new-ami", "ami-456")
			},
			setupMock: func(m *apitypes.MockEC2Client) {
				m.Instances = []types.Instance{
					{
						InstanceId: aws.String("i-123"),
						State: &types.InstanceState{
							Name: types.InstanceStateNameRunning,
						},
					},
				}
				m.StopInstancesError = fmt.Errorf("failed to stop instance")
			},
			wantErr: true,
			errMsg:  "failed to stop instance",
		},
		{
			name: "successful migration with enabled flag",
			setupCmd: func(cmd *cobra.Command) {
				cmd.Flags().Set("enabled", "true")
				cmd.Flags().Set("new-ami", "ami-456")
			},
			setupMock: func(m *apitypes.MockEC2Client) {
				m.Instances = []types.Instance{
					{
						InstanceId: aws.String("i-123"),
						State: &types.InstanceState{
							Name: types.InstanceStateNameRunning,
						},
						Tags: []types.Tag{
							{
								Key:   aws.String("ami-migrate"),
								Value: aws.String("enabled"),
							},
						},
					},
				}
				// Pre-configure instance states for the waiter
				m.SetInstanceState("i-123", types.InstanceStateNameRunning)
			},
		},
		{
			name: "successful migration with instance",
			setupCmd: func(cmd *cobra.Command) {
				cmd.Flags().Set("instance-id", "i-123")
				cmd.Flags().Set("new-ami", "ami-456")
			},
			setupMock: func(m *apitypes.MockEC2Client) {
				instance := types.Instance{
					InstanceId:   aws.String("i-123"),
					InstanceType: types.InstanceTypeT2Micro,
					ImageId:      aws.String("ami-123"), // Current AMI
					State: &types.InstanceState{
						Name: types.InstanceStateNameRunning,
					},
					BlockDeviceMappings: []types.InstanceBlockDeviceMapping{
						{
							DeviceName: aws.String("/dev/xvda"),
							Ebs: &types.EbsInstanceBlockDevice{
								VolumeId: aws.String("vol-123"),
							},
						},
					},
					Tags: []types.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String("test-instance"),
						},
					},
				}
				m.Instances = []types.Instance{instance}
				m.SetInstanceState("i-123", types.InstanceStateNameRunning)

				// Set up DescribeInstances response
				m.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{instance},
						},
					},
				}

				// Set up RunInstances response
				newInstance := instance
				newInstance.InstanceId = aws.String("i-456")
				newInstance.ImageId = aws.String("ami-456")
				m.RunInstancesOutput = &ec2.RunInstancesOutput{
					Instances: []types.Instance{newInstance},
				}

				// Set up TerminateInstances response
				m.TerminateInstancesOutput = &ec2.TerminateInstancesOutput{
					TerminatingInstances: []types.InstanceStateChange{
						{
							CurrentState: &types.InstanceState{
								Name: types.InstanceStateNameShuttingDown,
							},
							InstanceId: aws.String("i-123"),
							PreviousState: &types.InstanceState{
								Name: types.InstanceStateNameRunning,
							},
						},
					},
				}

				// Set up CreateSnapshot response
				m.CreateSnapshotOutput = &ec2.CreateSnapshotOutput{
					SnapshotId: aws.String("snap-123"),
				}

				// Set up CreateTags response
				m.CreateTagsOutput = &ec2.CreateTagsOutput{}

				// Set up StopInstances response
				m.StopInstancesOutput = &ec2.StopInstancesOutput{
					StoppingInstances: []types.InstanceStateChange{
						{
							CurrentState: &types.InstanceState{
								Name: types.InstanceStateNameStopping,
							},
							InstanceId: aws.String("i-123"),
							PreviousState: &types.InstanceState{
								Name: types.InstanceStateNameRunning,
							},
						},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new mock client and command for each test
			cmd, mockClient := setupTest("migrate", tt.setupMock)

			// Add migrate-specific flags
			cmd.Flags().String("new-ami", "", "ID of the new AMI to migrate to")

			// Set up command
			if tt.setupCmd != nil {
				tt.setupCmd(cmd)
			}

			// Set up command execution
			cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
				// Validate required flags
				instanceID, _ := cmd.Flags().GetString("instance-id")
				enabled, _ := cmd.Flags().GetBool("enabled")
				newAMI, _ := cmd.Flags().GetString("new-ami")

				if instanceID == "" && !enabled {
					return fmt.Errorf("either --instance-id or --enabled flag must be specified")
				}

				if newAMI == "" {
					return fmt.Errorf("--new-ami flag must be specified")
				}

				return nil
			}

			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				instanceID, _ := cmd.Flags().GetString("instance-id")
				enabled, _ := cmd.Flags().GetBool("enabled")
				newAMI, _ := cmd.Flags().GetString("new-ami")

				// Create AMI service
				svc := ami.NewService(client.GetEC2Client())

				// Get instances to migrate
				var instances []string
				if instanceID != "" {
					instances = []string{instanceID}
				} else if enabled {
					// Get all instances with ami-migrate=enabled tag
					taggedInstances, err := svc.ListUserInstances(context.Background(), "ami-migrate")
					if err != nil {
						return fmt.Errorf("failed to list instances: %v", err)
					}
					for _, instance := range taggedInstances {
						instances = append(instances, instance.InstanceID)
					}
				}

				if len(instances) == 0 {
					return fmt.Errorf("instance not found")
				}

				// Migrate each instance
				for _, instance := range instances {
					if err := svc.MigrateInstance(context.Background(), instance, newAMI); err != nil {
						return fmt.Errorf("failed to migrate instance %s: %v", instance, err)
					}
					logger.Info("Successfully migrated instance", "instanceID", instance)
				}

				return nil
			}

			// Execute command
			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
