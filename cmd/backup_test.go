package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/taemon1337/ec-manager/pkg/mock"
)

func NewBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup an EC2 instance",
		Long:  "Create a snapshot of an EC2 instance's root volume",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			instanceID, err := cmd.Flags().GetString("instance")
			if err != nil {
				return err
			}

			if instanceID == "" {
				return fmt.Errorf("instance ID must be set")
			}

			mockClient, ok := ctx.Value(mock.EC2ClientKey).(*mock.MockEC2Client)
			if !ok {
				return fmt.Errorf("failed to get EC2 client")
			}

			// Get instance details
			output, err := mockClient.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
				InstanceIds: []string{instanceID},
			})
			if err != nil {
				return fmt.Errorf("failed to get instance OS: %v", err)
			}

			if len(output.Reservations) == 0 || len(output.Reservations[0].Instances) == 0 {
				return fmt.Errorf("instance not found")
			}

			instance := output.Reservations[0].Instances[0]
			if len(instance.BlockDeviceMappings) == 0 {
				return fmt.Errorf("no block devices found")
			}

			volumeID := instance.BlockDeviceMappings[0].Ebs.VolumeId

			// Create snapshot
			_, err = mockClient.CreateSnapshot(ctx, &ec2.CreateSnapshotInput{
				VolumeId: volumeID,
			})
			if err != nil {
				return fmt.Errorf("failed to create snapshot: %v", err)
			}

			return nil
		},
	}

	cmd.Flags().String("instance", "", "Instance ID to backup")
	return cmd
}

func TestBackupCmd(t *testing.T) {
	mockClient := mock.NewMockEC2Client()
	instance := ec2types.Instance{
		InstanceId: aws.String("i-123"),
		ImageId:    aws.String("ami-original"),
		Platform:   ec2types.PlatformValues("windows"),
		BlockDeviceMappings: []ec2types.InstanceBlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/xvda"),
				Ebs: &ec2types.EbsInstanceBlockDevice{
					VolumeId: aws.String("vol-123"),
				},
			},
		},
	}

	instanceNoDevices := ec2types.Instance{
		InstanceId: aws.String("i-123"),
		ImageId:    aws.String("ami-original"),
		Platform:   ec2types.PlatformValues("windows"),
	}

	tests := []struct {
		name       string
		instance   string
		setupMock  func(*mock.MockEC2Client)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:     "success",
			instance: "i-123",
			setupMock: func(m *mock.MockEC2Client) {
				m.On("DescribeInstances", mock.Anything, &ec2.DescribeInstancesInput{
					InstanceIds: []string{"i-123"},
				}).Return(&ec2.DescribeInstancesOutput{
					Reservations: []ec2types.Reservation{
						{
							Instances: []ec2types.Instance{instance},
						},
					},
				}, nil).Once()

				m.On("CreateSnapshot", mock.Anything, &ec2.CreateSnapshotInput{
					VolumeId: aws.String("vol-123"),
				}).Return(&ec2.CreateSnapshotOutput{}, nil).Once()
			},
			wantErr: false,
		},
		{
			name:     "instance not found",
			instance: "i-999",
			setupMock: func(m *mock.MockEC2Client) {
				m.On("DescribeInstances", mock.Anything, &ec2.DescribeInstancesInput{
					InstanceIds: []string{"i-999"},
				}).Return(&ec2.DescribeInstancesOutput{
					Reservations: []ec2types.Reservation{},
				}, nil).Once()
			},
			wantErr:    true,
			wantErrMsg: "instance not found",
		},
		{
			name:     "describe_instances_error",
			instance: "i-123",
			setupMock: func(m *mock.MockEC2Client) {
				m.On("DescribeInstances", mock.Anything, &ec2.DescribeInstancesInput{
					InstanceIds: []string{"i-123"},
				}).Return(&ec2.DescribeInstancesOutput{}, fmt.Errorf("failed to describe instances")).Once()
			},
			wantErr:    true,
			wantErrMsg: "failed to get instance OS: failed to describe instances",
		},
		{
			name:     "create_snapshot_error",
			instance: "i-123",
			setupMock: func(m *mock.MockEC2Client) {
				m.On("DescribeInstances", mock.Anything, &ec2.DescribeInstancesInput{
					InstanceIds: []string{"i-123"},
				}).Return(&ec2.DescribeInstancesOutput{
					Reservations: []ec2types.Reservation{
						{
							Instances: []ec2types.Instance{instance},
						},
					},
				}, nil).Once()

				m.On("CreateSnapshot", mock.Anything, &ec2.CreateSnapshotInput{
					VolumeId: aws.String("vol-123"),
				}).Return(&ec2.CreateSnapshotOutput{}, fmt.Errorf("failed to create snapshot")).Once()
			},
			wantErr:    true,
			wantErrMsg: "failed to create snapshot: failed to create snapshot",
		},
		{
			name:     "no_block_devices",
			instance: "i-123",
			setupMock: func(m *mock.MockEC2Client) {
				m.On("DescribeInstances", mock.Anything, &ec2.DescribeInstancesInput{
					InstanceIds: []string{"i-123"},
				}).Return(&ec2.DescribeInstancesOutput{
					Reservations: []ec2types.Reservation{
						{
							Instances: []ec2types.Instance{instanceNoDevices},
						},
					},
				}, nil).Once()
			},
			wantErr:    true,
			wantErrMsg: "no block devices found",
		},
		{
			name:     "describe_instances_error_with_block_devices",
			instance: "i-123",
			setupMock: func(m *mock.MockEC2Client) {
				m.On("DescribeInstances", mock.Anything, &ec2.DescribeInstancesInput{
					InstanceIds: []string{"i-123"},
				}).Return(&ec2.DescribeInstancesOutput{
					Reservations: []ec2types.Reservation{
						{
							Instances: []ec2types.Instance{instanceNoDevices},
						},
					},
				}, fmt.Errorf("failed to describe instances")).Once()
			},
			wantErr:    true,
			wantErrMsg: "failed to get instance OS: failed to describe instances",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mockClient)

			cmd := NewBackupCmd()
			cmd.SetArgs([]string{"--instance", tt.instance})

			ctx := context.WithValue(context.Background(), mock.EC2ClientKey, mockClient)
			cmd.SetContext(ctx)

			err := cmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}
