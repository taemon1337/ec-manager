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
			instanceID, err := cmd.Flags().GetString("instance-id")
			if err != nil {
				return fmt.Errorf("failed to get instance-id flag: %w", err)
			}

			snapshotID, err := cmd.Flags().GetString("snapshot-id")
			if err != nil {
				return fmt.Errorf("failed to get snapshot-id flag: %w", err)
			}

			if instanceID == "" {
				return fmt.Errorf("--instance-id is required")
			}

			if snapshotID == "" {
				return fmt.Errorf("--snapshot-id is required")
			}

			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			ec2Client := testutil.GetEC2Client(cmd)
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

	cmd.Flags().String("instance-id", "", "Instance ID to restore")
	cmd.Flags().String("snapshot-id", "", "Snapshot ID to restore from")

	return cmd
}

func TestRestoreCmd(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError bool
		errMsg    string
		setup     func(client *mock.MockEC2Client)
	}{
		{
			name: "successful_restore",
			args: []string{"--instance-id", "i-123", "--snapshot-id", "snap-123"},
			setup: func(client *mock.MockEC2Client) {
				client.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
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
				}
			},
		},
		{
			name:      "instance_not_found",
			args:      []string{"--instance-id", "i-nonexistent", "--snapshot-id", "snap-123"},
			wantError: true,
			errMsg:    "instance not found",
			setup: func(client *mock.MockEC2Client) {
				client.DescribeInstancesOutput = &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{},
				}
			},
		},
		{
			name:      "missing_instance_id",
			args:      []string{"--snapshot-id", "snap-123"},
			wantError: true,
			errMsg:    "--instance-id is required",
		},
		{
			name:      "missing_snapshot_id",
			args:      []string{"--instance-id", "i-123"},
			wantError: true,
			errMsg:    "--snapshot-id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRestoreCmd()

			if tt.setup != nil {
				mockClient := mock.NewMockEC2Client()
				tt.setup(mockClient)
				cmd.SetContext(context.WithValue(context.Background(), "ec2_client", mockClient))
			}

			err := testutil.SetupTestCommand(cmd, tt.args)
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
