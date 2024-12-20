package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/client"
	"github.com/taemon1337/ec-manager/pkg/logger"
	"github.com/taemon1337/ec-manager/pkg/testutil"
	apitypes "github.com/taemon1337/ec-manager/pkg/types"
	"github.com/spf13/cobra"
)

func TestBackupCmd(t *testing.T) {
	// Initialize test logger
	testutil.InitTestLogger(t)

	// Store original values
	originalInstanceID := instanceID
	originalEnabled := enabled
	originalLogLevel := logLevel

	// Reset flags after tests
	t.Cleanup(func() {
		instanceID = originalInstanceID
		enabled = originalEnabled
		logLevel = originalLogLevel
	})

	// Initialize logger for tests with debug level
	logLevel = "debug"
	logger.Init(logger.LogLevel(logLevel))

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
		setupMock   func(*apitypes.MockEC2Client)
	}{
		{
			name: "successful backup",
			args: []string{"--instance-id", "i-123"},
			setupMock: func(m *apitypes.MockEC2Client) {
				// Set up the instance
				m.Instance = &types.Instance{
					InstanceId: aws.String("i-123"),
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
				}
				m.Instances = []types.Instance{*m.Instance}

				// Set up the describe instances response
				m.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: m.Instances,
						},
					},
				}

				// Set up the create snapshot response
				m.CreateSnapshotOutput = &ec2.CreateSnapshotOutput{
					SnapshotId: aws.String("snap-123"),
				}
			},
			wantErr: false,
		},
		{
			name: "missing required flags",
			args: []string{},
			setupMock: func(m *apitypes.MockEC2Client) {},
			wantErr: true,
			errContains: "required flag(s) \"instance-id\" not set",
		},
		{
			name: "backup failure",
			args: []string{"--instance-id", "i-123"},
			setupMock: func(m *apitypes.MockEC2Client) {
				// Set up the instance
				m.Instance = &types.Instance{
					InstanceId: aws.String("i-123"),
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
				}
				m.Instances = []types.Instance{*m.Instance}

				// Set up the describe instances response
				m.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: m.Instances,
						},
					},
				}

				// Set up the create snapshot to fail
				m.CreateSnapshotOutput = nil // Ensure output is nil when error is set
				m.CreateSnapshotError = fmt.Errorf("failed to create snapshot")
			},
			wantErr: true,
			errContains: "failed to backup instance",
		},
		{
			name: "successful backup with enabled flag",
			args: []string{"--enabled"},
			setupMock: func(m *apitypes.MockEC2Client) {
				// Set up the instance
				m.Instance = &types.Instance{
					InstanceId: aws.String("i-123"),
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
							Key:   aws.String("ami-migrate"),
							Value: aws.String("enabled"),
						},
					},
				}
				m.Instances = []types.Instance{*m.Instance}

				// Set up the describe instances response
				m.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: m.Instances,
						},
					},
				}

				// Set up the create snapshot response
				m.CreateSnapshotOutput = &ec2.CreateSnapshotOutput{
					SnapshotId: aws.String("snap-123"),
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test command
			cmd, _ := setupTest("backup", tt.setupMock)

			// Add backup-specific flags
			cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
				instanceID, _ = cmd.Flags().GetString("instance-id")
				enabled, _ = cmd.Flags().GetBool("enabled")
				if instanceID == "" && !enabled {
					return fmt.Errorf("required flag(s) \"instance-id\" not set")
				}
				return nil
			}

			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				// Get flag values
				instanceID, _ := cmd.Flags().GetString("instance-id")
				enabled, _ := cmd.Flags().GetBool("enabled")

				// Create EC2 client
				ec2Client, err := client.GetEC2Client(cmd.Context())
				if err != nil {
					return fmt.Errorf("failed to get EC2 client: %w", err)
				}

				// Create AMI service
				svc := ami.NewService(ec2Client)

				// Get instances to backup
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
					return fmt.Errorf("no instances found to backup")
				}

				// Backup each instance
				for _, instance := range instances {
					if err := svc.BackupInstance(context.Background(), instance); err != nil {
						return fmt.Errorf("failed to backup instance %s: %v", instance, err)
					}
					logger.Info("Successfully backed up instance", "instanceID", instance)
				}

				return nil
			}

			// Set command args
			cmd.SetArgs(tt.args)

			// Execute command
			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
