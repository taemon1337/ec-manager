package cmd

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/taemon1337/ec-manager/pkg/client"
	"github.com/taemon1337/ec-manager/pkg/logger"
	ecTypes "github.com/taemon1337/ec-manager/pkg/types"
)

func setupBackupCommand(t *testing.T, setupMock func(*ecTypes.MockEC2Client)) *cobra.Command {
	// Initialize test logger
	logger.Init(logger.DebugLevel)

	// Create a new mock client
	mockClient := &ecTypes.MockEC2Client{}
	if setupMock != nil {
		setupMock(mockClient)
	}

	// Create client and set mock mode
	c := client.NewClient()
	c.SetMockMode(true)
	c.SetEC2Client(mockClient)
	awsClient = c

	// Create and setup command
	cmd := &cobra.Command{
		Use:   backupCmd.Use,
		Short: backupCmd.Short,
		PreRunE: backupCmd.PreRunE,
		RunE:  backupCmd.RunE,
	}
	cmd.Flags().String("instance-id", "", "ID of the instance")
	cmd.Flags().Bool("enabled", false, "Process all instances with ami-migrate=enabled tag")

	return cmd
}

func TestBackupCmd(t *testing.T) {
	t.Run("successful backup", func(t *testing.T) {
		cmd := setupBackupCommand(t, func(m *ecTypes.MockEC2Client) {
			m.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
				return &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceId: aws.String("i-123"),
									ImageId:    aws.String("ami-old"),
									State: &types.InstanceState{
										Name: types.InstanceStateNameRunning,
									},
								},
							},
						},
					},
				}, nil
			}
			m.CreateImageFunc = func(ctx context.Context, params *ec2.CreateImageInput, optFns ...func(*ec2.Options)) (*ec2.CreateImageOutput, error) {
				return &ec2.CreateImageOutput{
					ImageId: aws.String("ami-backup"),
				}, nil
			}
			m.DescribeImagesFunc = func(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
				return &ec2.DescribeImagesOutput{
					Images: []types.Image{
						{
							ImageId: aws.String("ami-backup"),
							State:   types.ImageStateAvailable,
						},
					},
				}, nil
			}
			m.CreateTagsFunc = func(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
				return &ec2.CreateTagsOutput{}, nil
			}
		})

		cmd.SetArgs([]string{"--instance-id", "i-123"})
		err := cmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("instance not found", func(t *testing.T) {
		cmd := setupBackupCommand(t, func(m *ecTypes.MockEC2Client) {
			m.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
				return &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
				}, nil
			}
		})

		cmd.SetArgs([]string{"--instance-id", "i-nonexistent"})
		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "instance not found")
	})

	tests := []struct {
		name      string
		args      []string
		setupMock func(*ecTypes.MockEC2Client)
		wantErr   bool
	}{
		{
			name: "instance stopped",
			args: []string{"--instance-id", "i-123"},
			setupMock: func(m *ecTypes.MockEC2Client) {
				m.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
					return &ec2.DescribeInstancesOutput{
						Reservations: []types.Reservation{
							{
								Instances: []types.Instance{
									{
										InstanceId: aws.String("i-123"),
										State: &types.InstanceState{
											Name: types.InstanceStateNameStopped,
										},
									},
								},
							},
						},
					}, nil
				}
				m.CreateImageFunc = func(ctx context.Context, params *ec2.CreateImageInput, optFns ...func(*ec2.Options)) (*ec2.CreateImageOutput, error) {
					return &ec2.CreateImageOutput{
						ImageId: aws.String("ami-backup"),
					}, nil
				}
				m.DescribeImagesFunc = func(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
					return &ec2.DescribeImagesOutput{
						Images: []types.Image{
							{
								ImageId: aws.String("ami-backup"),
								State:   types.ImageStateAvailable,
							},
						},
					}, nil
				}
				m.CreateTagsFunc = func(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
					return &ec2.CreateTagsOutput{}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "no instance ID and enabled flag not set",
			args: []string{},
			setupMock: func(m *ecTypes.MockEC2Client) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := setupBackupCommand(t, tt.setupMock)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
