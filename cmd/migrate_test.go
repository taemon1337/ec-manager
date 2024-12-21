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
	"github.com/taemon1337/ec-manager/pkg/ami"
	"github.com/taemon1337/ec-manager/pkg/client"
	"github.com/taemon1337/ec-manager/pkg/logger"
	"github.com/taemon1337/ec-manager/pkg/testutil"
	apitypes "github.com/taemon1337/ec-manager/pkg/types"
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
				m.InstanceStates = make(map[string]types.InstanceStateName)
				m.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId:   aws.String("i-123"),
									InstanceType: types.InstanceTypeT2Micro,
									ImageId:      aws.String("ami-123"),
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
								},
							},
						},
					},
				}
				m.RunInstancesOutput = &ec2.RunInstancesOutput{
					Instances: []types.Instance{
						{
							InstanceId: aws.String("i-456"),
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
				m.InstanceStates = make(map[string]types.InstanceStateName)
				m.InstanceStates["i-123"] = types.InstanceStateNameRunning
				m.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId:   aws.String("i-123"),
									ImageId:      aws.String("ami-123"),
									InstanceType: types.InstanceTypeT2Micro,
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
								},
							},
						},
					},
				}
				m.StopInstancesError = fmt.Errorf("failed to stop instance")
				m.Instance = &types.Instance{
					InstanceId:   aws.String("i-123"),
					ImageId:      aws.String("ami-123"),
					InstanceType: types.InstanceTypeT2Micro,
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
			},
			validate: func(t *testing.T, m *apitypes.MockEC2Client) {
				// Instance should still be running since stop failed
				assert.Equal(t, types.InstanceStateNameRunning, m.InstanceStates["i-123"], "instance should still be running after failed stop")
			},
			wantErr: true,
			errMsg:  "failed to stop instance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create command
			cmd := NewMigrateCmd()
			cmd.SilenceUsage = true
			if tt.setupCmd != nil {
				tt.setupCmd(cmd)
			}

			// Create mock client
			mockClient := &apitypes.MockEC2Client{
				InstanceStates: make(map[string]types.InstanceStateName),
			}
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			// Set mock client
			if err := client.SetEC2Client(mockClient); err != nil {
				t.Fatal(err)
			}

			// Run command
			err := cmd.Execute()

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			// Run validation
			if tt.validate != nil {
				tt.validate(t, mockClient)
			}
		})
	}
}

func NewMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate instances to a new AMI",
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceID, _ := cmd.Flags().GetString("instance-id")
			enabled, _ := cmd.Flags().GetBool("enabled")
			newAMI, _ := cmd.Flags().GetString("new-ami")

			// Validate required flags
			if instanceID == "" && !enabled {
				return fmt.Errorf("either --instance-id or --enabled flag must be specified")
			}

			if newAMI == "" {
				return fmt.Errorf("--new-ami flag must be specified")
			}

			// Create EC2 client
			ec2Client, err := client.GetEC2Client(cmd.Context())
			if err != nil {
				return err
			}

			// Create AMI service
			svc := ami.NewService(ec2Client)

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
		},
	}

	cmd.Flags().String("instance-id", "", "ID of the instance to migrate")
	cmd.Flags().Bool("enabled", false, "Migrate all instances with ami-migrate=enabled tag")
	cmd.Flags().String("new-ami", "", "ID of the new AMI to migrate to")

	return cmd
}
