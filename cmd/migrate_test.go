package cmd

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/taemon1337/ec-manager/pkg/client"
	"github.com/taemon1337/ec-manager/pkg/logger"
	ecTypes "github.com/taemon1337/ec-manager/pkg/types"
	"context"
	"strings"
)

func setupMigrateCommand(t *testing.T, setupMock func(*ecTypes.MockEC2Client)) *cobra.Command {
	// Initialize test logger
	logger.Init(logger.DebugLevel)

	// Create mock EC2 client
	mockClient := &ecTypes.MockEC2Client{}
	if setupMock != nil {
		setupMock(mockClient)
	}

	// Enable mock mode and set mock client
	c := client.NewClient()
	c.SetMockMode(true)
	c.SetEC2Client(mockClient)
	awsClient = c

	// Create command
	cmd := &cobra.Command{
		Use:     migrateCmd.Use,
		Short:   migrateCmd.Short,
		Long:    migrateCmd.Long,
		PreRunE: migrateCmd.PreRunE,
		RunE:    migrateCmd.RunE,
	}

	// Add flags
	cmd.Flags().String("instance-id", "", "ID of the instance to migrate")
	cmd.Flags().String("new-ami", "", "ID of the new AMI to migrate to")
	cmd.Flags().Bool("enabled", false, "Migrate all instances with ami-migrate=enabled tag")

	return cmd
}

func TestMigrateCmd(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		setupMock    func(*ecTypes.MockEC2Client)
		wantErr      bool
		expectedErr  string
	}{
		{
			name: "successful migration",
			args: []string{"--instance-id", "i-123", "--new-ami", "ami-new"},
			setupMock: func(m *ecTypes.MockEC2Client) {
				// Track instance state for i-123
				var instanceState types.InstanceStateName = types.InstanceStateNameRunning

				// Mock DescribeInstances to return current instance state
				m.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
					if len(params.InstanceIds) > 0 && params.InstanceIds[0] == "i-123" {
						return &ec2.DescribeInstancesOutput{
							Reservations: []types.Reservation{
								{
									Instances: []types.Instance{
										{
											InstanceId: aws.String("i-123"),
											ImageId:    aws.String("ami-old"),
											State: &types.InstanceState{
												Name: instanceState,
											},
											BlockDeviceMappings: []types.InstanceBlockDeviceMapping{
												{
													Ebs: &types.EbsInstanceBlockDevice{
														VolumeId: aws.String("vol-123"),
													},
												},
											},
											InstanceType: types.InstanceTypeT2Micro,
										},
									},
								},
							},
						}, nil
					}
					return &ec2.DescribeInstancesOutput{}, nil
				}

				// Mock StopInstances to update instance state
				m.StopInstancesFunc = func(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
					instanceState = types.InstanceStateNameStopped
					return &ec2.StopInstancesOutput{
						StoppingInstances: []types.InstanceStateChange{
							{
								CurrentState: &types.InstanceState{
									Name: types.InstanceStateNameStopped,
								},
								InstanceId: aws.String("i-123"),
							},
						},
					}, nil
				}

				// Mock CreateSnapshot
				m.CreateSnapshotFunc = func(ctx context.Context, params *ec2.CreateSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotOutput, error) {
					return &ec2.CreateSnapshotOutput{
						SnapshotId: aws.String("snap-123"),
					}, nil
				}

				// Mock RunInstances
				m.RunInstancesFunc = func(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
					return &ec2.RunInstancesOutput{
						Instances: []types.Instance{
							{
								InstanceId: aws.String("i-new"),
								State: &types.InstanceState{
									Name: types.InstanceStateNamePending,
								},
							},
						},
					}, nil
				}

				// Mock TerminateInstances
				m.TerminateInstancesFunc = func(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
					instanceState = types.InstanceStateNameTerminated
					return &ec2.TerminateInstancesOutput{
						TerminatingInstances: []types.InstanceStateChange{
							{
								CurrentState: &types.InstanceState{
									Name: types.InstanceStateNameShuttingDown,
								},
								InstanceId: aws.String("i-123"),
							},
						},
					}, nil
				}

				// Mock CreateTags
				m.CreateTagsFunc = func(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
					return &ec2.CreateTagsOutput{}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "instance not found",
			args: []string{"--instance-id", "i-nonexistent", "--new-ami", "ami-new"},
			setupMock: func(m *ecTypes.MockEC2Client) {
				m.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
					return &ec2.DescribeInstancesOutput{
						Reservations: []types.Reservation{},
					}, nil
				}
			},
			wantErr: true,
			expectedErr: "failed to migrate instance i-nonexistent: get instance: instance not found: i-nonexistent",
		},
		{
			name: "no instance ID and enabled flag not set",
			args: []string{},
			setupMock: func(m *ecTypes.MockEC2Client) {},
			wantErr: true,
			expectedErr: `required flag(s) "instance-id" or "enabled" not set`,
		},
		{
			name: "missing new-ami flag",
			args: []string{"--instance-id", "i-123"},
			setupMock: func(m *ecTypes.MockEC2Client) {},
			wantErr: true,
			expectedErr: "--new-ami flag must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := setupMigrateCommand(t, tt.setupMock)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					// Strip the "Error: " prefix that Cobra adds
					errStr := err.Error()
					if strings.HasPrefix(errStr, "Error: ") {
						errStr = strings.TrimPrefix(errStr, "Error: ")
					}
					assert.Equal(t, tt.expectedErr, errStr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
