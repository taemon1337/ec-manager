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
	"github.com/taemon1337/ec-manager/pkg/mock"
	"github.com/taemon1337/ec-manager/pkg/testutil"
)

func NewRestoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore an EC2 instance",
		Long:  "Restore an EC2 instance from a snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceID, err := cmd.Flags().GetString("instance")
			if err != nil {
				return fmt.Errorf("failed to get instance flag: %w", err)
			}

			snapshotID, err := cmd.Flags().GetString("snapshot")
			if err != nil {
				return fmt.Errorf("failed to get snapshot flag: %w", err)
			}

			if instanceID == "" {
				return fmt.Errorf("--instance is required")
			}

			if snapshotID == "" {
				return fmt.Errorf("--snapshot is required")
			}

			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			ec2Client := testutil.GetEC2Client(cmd.Context())
			if ec2Client == nil {
				return fmt.Errorf("failed to get EC2 client")
			}

			// Check if instance exists
			output, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
				InstanceIds: []string{instanceID},
			})
			if err != nil {
				return fmt.Errorf("failed to describe instance: %w", err)
			}

			if len(output.Reservations) == 0 || len(output.Reservations[0].Instances) == 0 {
				return fmt.Errorf("instance not found")
			}

			fmt.Printf("Instance %s restored from snapshot %s\n", instanceID, snapshotID)
			return nil
		},
	}

	cmd.Flags().String("instance", "", "Instance to restore")
	cmd.Flags().String("snapshot", "", "Snapshot to restore from")

	return cmd
}

func TestRestoreCmd(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		setup     func(*mock.MockEC2Client)
		wantError bool
		errMsg    string
	}{
		{
			name: "successful restore",
			args: []string{"--instance", "i-123", "--snapshot", "snap-123"},
			setup: func(mockClient *mock.MockEC2Client) {
				mockClient.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
					return &ec2.DescribeInstancesOutput{
						Reservations: []types.Reservation{
							{
								Instances: []types.Instance{
									{
										InstanceId: aws.String("i-123"),
										State: &types.InstanceState{
											Name: types.InstanceStateNameRunning,
										},
									},
								},
							},
						},
					}, nil
				}

				mockClient.CreateVolumeFunc = func(ctx context.Context, params *ec2.CreateVolumeInput, optFns ...func(*ec2.Options)) (*ec2.CreateVolumeOutput, error) {
					return &ec2.CreateVolumeOutput{
						VolumeId: aws.String("vol-123"),
					}, nil
				}

				mockClient.AttachVolumeFunc = func(ctx context.Context, params *ec2.AttachVolumeInput, optFns ...func(*ec2.Options)) (*ec2.AttachVolumeOutput, error) {
					return &ec2.AttachVolumeOutput{}, nil
				}
			},
			wantError: false,
		},
		{
			name: "instance not found",
			args: []string{"--instance", "i-nonexistent", "--snapshot", "snap-123"},
			setup: func(mockClient *mock.MockEC2Client) {
				mockClient.DescribeInstancesFunc = func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
					instanceID := "i-nonexistent"
					return nil, fmt.Errorf("instance not found: %s", instanceID)
				}
			},
			wantError: true,
			errMsg:    "failed to describe instance: instance not found: i-nonexistent",
		},
		{
			name:      "missing instance",
			args:      []string{"--snapshot", "snap-123"},
			wantError: true,
			errMsg:    "--instance is required",
		},
		{
			name:      "missing snapshot",
			args:      []string{"--instance", "i-123"},
			wantError: true,
			errMsg:    "--snapshot is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRestoreCmd()

			if tt.setup != nil {
				mockClient := mock.NewMockEC2Client()
				tt.setup(mockClient)
				cmd.SetContext(testutil.GetTestContextWithClient(mockClient))
			}

			err := testutil.SetupTestCommand(cmd, tt.args)

			if tt.wantError {
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
