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

	tests := []struct {
		name      string
		setupCmd  func(*cobra.Command)
		setupMock func(*apitypes.MockEC2Client)
		validate  func(*testing.T, *apitypes.MockEC2Client)
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
								Name: types.InstanceStateNameStopped,
							},
							InstanceId: aws.String("i-123"),
							PreviousState: &types.InstanceState{
								Name: types.InstanceStateNameRunning,
							},
						},
					},
				}

				// Set up StartInstances response
				m.StartInstancesOutput = &ec2.StartInstancesOutput{
					StartingInstances: []types.InstanceStateChange{
						{
							CurrentState: &types.InstanceState{
								Name: types.InstanceStateNameRunning,
							},
							InstanceId: aws.String("i-456"),
							PreviousState: &types.InstanceState{
								Name: types.InstanceStateNameStopped,
							},
						},
					},
				}
			},
			validate: func(t *testing.T, m *apitypes.MockEC2Client) {
				// Verify instance state transitions
				assert.Equal(t, types.InstanceStateNameTerminated, m.GetInstanceState("i-123"), "original instance should be terminated")
				assert.Equal(t, types.InstanceStateNameRunning, m.GetInstanceState("i-456"), "new instance should be running")
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
				cmd.Flags().Set("instance-id", "i-nonexistent")
				cmd.Flags().Set("new-ami", "ami-456")
			},
			setupMock: func(m *apitypes.MockEC2Client) {
				m.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
				}
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
				instance := types.Instance{
					InstanceId: aws.String("i-123"),
					ImageId:    aws.String("ami-123"),
					State: &types.InstanceState{
						Name: types.InstanceStateNameRunning,
					},
				}
				m.Instances = []types.Instance{instance}
				m.SetInstanceState("i-123", types.InstanceStateNameRunning)
				m.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{instance},
						},
					},
				}
				m.StopInstancesError = fmt.Errorf("failed to stop instance")
			},
			validate: func(t *testing.T, m *apitypes.MockEC2Client) {
				// Instance should still be running since stop failed
				assert.Equal(t, types.InstanceStateNameRunning, m.GetInstanceState("i-123"), "instance should still be running after failed stop")
			},
			wantErr: true,
			errMsg:  "failed to stop instance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new command for each test
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

			// Run validation if provided
			if tt.validate != nil {
				tt.validate(t, mockClient)
			}
		})
	}
}
