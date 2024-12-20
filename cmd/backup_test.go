package cmd

import (
	"bytes"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/taemon1337/ami-migrate/pkg/ami/mocks"
	"github.com/taemon1337/ami-migrate/pkg/client"
	"github.com/stretchr/testify/assert"
	"fmt"
)

func TestBackupCmd(t *testing.T) {
	tests := []struct {
		name        string
		mockClient  *mocks.MockEC2Client
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name: "successful backup - no specific instance",
			mockClient: &mocks.MockEC2Client{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId: aws.String("i-123"),
									State: &types.InstanceState{
										Name: types.InstanceStateNameRunning,
									},
									BlockDeviceMappings: []types.InstanceBlockDeviceMapping{
										{
											DeviceName: aws.String("/dev/sda1"),
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
								},
							},
						},
					},
				},
				CreateSnapshotOutput: &ec2.CreateSnapshotOutput{
					SnapshotId: aws.String("snap-123"),
				},
				CreateTagsOutput: &ec2.CreateTagsOutput{},
			},
			args:    []string{"backup"},
			wantErr: false,
		},
		{
			name: "successful backup - specific instance",
			mockClient: &mocks.MockEC2Client{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId: aws.String("i-123"),
									State: &types.InstanceState{
										Name: types.InstanceStateNameRunning,
									},
									BlockDeviceMappings: []types.InstanceBlockDeviceMapping{
										{
											DeviceName: aws.String("/dev/sda1"),
											Ebs: &types.EbsInstanceBlockDevice{
												VolumeId: aws.String("vol-123"),
											},
										},
									},
								},
							},
						},
					},
				},
				CreateSnapshotOutput: &ec2.CreateSnapshotOutput{
					SnapshotId: aws.String("snap-123"),
				},
				CreateTagsOutput: &ec2.CreateTagsOutput{},
			},
			args:    []string{"backup", "--instance-id", "i-123"},
			wantErr: false,
		},
		{
			name: "error - instance not found",
			mockClient: &mocks.MockEC2Client{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
				},
			},
			args:        []string{"backup", "--instance-id", "i-nonexistent"},
			wantErr:     true,
			errContains: "no instances found",
		},
		{
			name: "error - describe instances error",
			mockClient: &mocks.MockEC2Client{
				DescribeInstancesError: fmt.Errorf("AWS API error"),
			},
			args:        []string{"backup", "--instance-id", "i-123"},
			wantErr:     true,
			errContains: "AWS API error",
		},
		{
			name: "error - create snapshot error",
			mockClient: &mocks.MockEC2Client{
				DescribeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId: aws.String("i-123"),
									State: &types.InstanceState{
										Name: types.InstanceStateNameRunning,
									},
									BlockDeviceMappings: []types.InstanceBlockDeviceMapping{
										{
											DeviceName: aws.String("/dev/sda1"),
											Ebs: &types.EbsInstanceBlockDevice{
												VolumeId: aws.String("vol-123"),
											},
										},
									},
								},
							},
						},
					},
				},
				CreateSnapshotError: fmt.Errorf("failed to create snapshot"),
			},
			args:        []string{"backup", "--instance-id", "i-123"},
			wantErr:     true,
			errContains: "failed to create snapshot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset client state before each test
			client.ResetClient()
			client.SetEC2Client(tt.mockClient)

			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()

			// Reset client state after test
			client.ResetClient()

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
